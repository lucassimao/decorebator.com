package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

func renderAssets(cfg *Config, generator ImageGenerator, resizeMode, provider string, creative int, storeFilter, localeFilter, slotFilter string, attempts int, invertMask, dryRun, overwrite bool) error {
	if attempts <= 0 {
		return fmt.Errorf("attempts must be >= 1")
	}

	stores, err := selectStores(cfg, storeFilter)
	if err != nil {
		return err
	}

	locales := selectLocales(cfg, localeFilter)
	if len(locales) == 0 {
		return fmt.Errorf("no locales matched: %s", localeFilter)
	}

	slots := selectSlots(cfg, slotFilter)
	if len(slots) == 0 {
		return fmt.Errorf("no slots matched: %s", slotFilter)
	}

	guidePath, maskPath, err := ensureGuideAndMask(cfg, invertMask, dryRun, overwrite)
	if err != nil {
		return err
	}
	fmt.Printf("generator: %s\n", generator.Name())
	fmt.Printf("resize mode: %s\n", resizeMode)
	concurrency := attempts
	if concurrency < 1 {
		concurrency = 1
	}
	fmt.Printf("concurrency: %d\n", concurrency)
	fmt.Printf("guide: %s\n", guidePath)
	fmt.Printf("mask: %s\n", maskPath)

	type job struct {
		store          string
		storeCfg       StoreConfig
		locale         string
		slotCfg        SlotConfig
		copyEntry      CopyEntry
		screenshotPath string
		framePath      string
		useFrame       bool
		attempt        int
		outputPath     string
	}

	var jobs []job
	for _, store := range stores {
		storeCfg := cfg.Stores[store]
		fmt.Printf("store: %s (%dx%d)\n", store, storeCfg.Width, storeCfg.Height)
		framePath := framePathForStore(cfg, store)
		useFrame := framePath != "" && fileExists(framePath)
		if framePath != "" && !useFrame {
			fmt.Printf("frame not found, skipping frame: %s\n", framePath)
		}
		for _, loc := range locales {
			copyFile, err := loadCopyFile(cfg, loc)
			if err != nil {
				return err
			}
			fmt.Printf("locale: %s\n", loc)
			for _, slotCfg := range slots {
				copyEntry, ok := copyFile[slotCfg.Key]
				if !ok {
					return fmt.Errorf("missing copy for slot %s in locale %s", slotCfg.Key, loc)
				}

				screenshotPath, err := findScreenshot(cfg, loc, slotCfg.Key)
				if err != nil {
					return err
				}
				fmt.Printf("slot: %s (screenshot: %s)\n", slotCfg.Key, screenshotPath)

				for attempt := 1; attempt <= attempts; attempt++ {
					outputPath := buildOutputPath(cfg, store, loc, slotCfg.Key, storeCfg.Format, attempts, attempt)
					jobs = append(jobs, job{
						store:          store,
						storeCfg:       storeCfg,
						locale:         loc,
						slotCfg:        slotCfg,
						copyEntry:      copyEntry,
						screenshotPath: screenshotPath,
						framePath:      framePath,
						useFrame:       useFrame,
						attempt:        attempt,
						outputPath:     outputPath,
					})
				}
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobCh := make(chan job)
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	worker := func() {
		defer wg.Done()
		for j := range jobCh {
			if ctx.Err() != nil {
				return
			}
			if !overwrite {
				if _, err := os.Stat(j.outputPath); err == nil {
					fmt.Printf("skip existing: %s\n", j.outputPath)
					continue
				}
			}
			if dryRun {
				fmt.Printf("render %s %s %s try %d -> %s\n", j.store, j.locale, j.slotCfg.Key, j.attempt, j.outputPath)
				continue
			}

			fmt.Printf("attempt %d/%d: prepare request\n", j.attempt, attempts)
			prompt := buildFullPrompt(cfg, provider, creative, j.store, j.storeCfg, j.slotCfg, j.copyEntry)
			img, err := generator.Generate(ImageRequest{
				Prompt:         prompt,
				BaseWidth:      cfg.Defaults.BaseWidth,
				BaseHeight:     cfg.Defaults.BaseHeight,
				Layout:         cfg.Layout,
				ScreenshotPath: j.screenshotPath,
				GuidePath:      guidePath,
				MaskPath:       maskPath,
				FramePath:      j.framePath,
				UseFrame:       j.useFrame,
			})
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				cancel()
				return
			}

			fmt.Printf("attempt %d/%d: resize + write\n", j.attempt, attempts)
			final := resizeForStore(img, resizeMode, j.store, j.storeCfg.Width, j.storeCfg.Height)
			if err := writeImage(j.outputPath, j.storeCfg.Format, final); err != nil {
				select {
				case errCh <- err:
				default:
				}
				cancel()
				return
			}
			fmt.Printf("wrote %s\n", j.outputPath)
		}
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker()
	}

	go func() {
		defer close(jobCh)
		for _, j := range jobs {
			if ctx.Err() != nil {
				return
			}
			jobCh <- j
		}
	}()

	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
	}

	return nil
}

func selectStores(cfg *Config, filter string) ([]string, error) {
	if filter == "all" {
		stores := make([]string, 0, len(cfg.Stores))
		for key := range cfg.Stores {
			stores = append(stores, key)
		}
		sort.Strings(stores)
		return stores, nil
	}
	if _, ok := cfg.Stores[filter]; !ok {
		return nil, fmt.Errorf("unknown store: %s", filter)
	}
	return []string{filter}, nil
}

func selectLocales(cfg *Config, filter string) []string {
	if filter == "all" {
		return cfg.Locales
	}
	for _, loc := range cfg.Locales {
		if loc == filter {
			return []string{loc}
		}
	}
	return nil
}

func selectSlots(cfg *Config, filter string) []SlotConfig {
	if filter == "all" {
		return cfg.Slots
	}
	for _, slot := range cfg.Slots {
		if slot.Key == filter || fmt.Sprintf("%d", slot.ID) == filter {
			return []SlotConfig{slot}
		}
	}
	return nil
}

func loadCopyFile(cfg *Config, locale string) (CopyFile, error) {
	path := filepath.Join(cfg.Defaults.CopyDir, locale+".json")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var copyFile CopyFile
	if err := json.Unmarshal(content, &copyFile); err != nil {
		return nil, err
	}
	return copyFile, nil
}

func findScreenshot(cfg *Config, locale, slot string) (string, error) {
	extensions := []string{".png", ".jpg", ".jpeg"}
	base := filepath.Join(cfg.Defaults.ScreenshotDir, locale, "iphone11", slot)
	for _, ext := range extensions {
		path := base + ext
		if fileExists(path) {
			return path, nil
		}
	}
	return "", fmt.Errorf("missing screenshot: %s(.png/.jpg/.jpeg)", base)
}

func buildOutputPath(cfg *Config, store, locale, slot, format string, attempts, attempt int) string {
	filename := slot + "." + format
	if attempts > 1 {
		filename = fmt.Sprintf("%s_try%d.%s", slot, attempt, format)
	}
	return filepath.Join(cfg.Defaults.OutputDir, store, locale, filename)
}

func framePathForStore(cfg *Config, store string) string {
	switch store {
	case "appstore":
		if cfg.Defaults.FramePathApp != "" {
			return cfg.Defaults.FramePathApp
		}
	case "play":
		if cfg.Defaults.FramePathPlay != "" {
			return cfg.Defaults.FramePathPlay
		}
	}
	return cfg.Defaults.FramePath
}

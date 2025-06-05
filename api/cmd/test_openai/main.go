package main

import (
	"fmt"
	"log"

	"decorebator.com/internal/openai"
)

func main() {
	// Test generating a definition for a verb
	token := "test"
	
	fmt.Printf("Testing ChatGPT definition generation for '%s'...\n", token)
	
	result, err := openai.GetDefinition(token)
	if err != nil {
		log.Fatal("Failed to get definition:", err)
	}
	
	fmt.Printf("Pronunciation: %s\n", result.Pronunciation)
	fmt.Printf("Number of definitions: %d\n", len(result.Definitions))
	
	for i, def := range result.Definitions {
		fmt.Printf("\nDefinition %d:\n", i+1)
		fmt.Printf("  Token: %s\n", def.Token)
		fmt.Printf("  Language: %s\n", def.Language)
		fmt.Printf("  PartOfSpeech: %s\n", def.PartOfSpeech)
		fmt.Printf("  Meaning: %s\n", def.Meaning)
		fmt.Printf("  Source: %s\n", def.Source)
		fmt.Printf("  Examples count: %d\n", len(def.Examples))
		
		if len(def.Examples) > 0 {
			fmt.Println("  Examples:")
			for j, ex := range def.Examples {
				fmt.Printf("    [%d] %s\n", j, ex)
			}
		}
		
		fmt.Printf("  Inflections count: %d\n", len(def.Inflections))
		if len(def.Inflections) > 0 {
			fmt.Println("  Inflections:")
			for j, inf := range def.Inflections {
				fmt.Printf("    [%d] %s (%s) - %d examples\n", j, inf.Inflection, inf.Tense, len(inf.Examples))
				for k, ex := range inf.Examples {
					fmt.Printf("      [%d] %s\n", k, ex)
				}
			}
		}
	}
}
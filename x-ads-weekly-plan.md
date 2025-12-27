# Decorebator — X Ads Weekly Plan (Image + Text)

Context agreed on Nov 2, 2025
- Cadence: target 1 country per week.
- Daily budget cap: R$ 100 (or equivalent currency in ad account).
- Objective: Website Traffic campaigns linking directly to App Store / Google Play pages.
- Primary KPI (simple start): outbound link clicks to the store pages (X Ads reporting). Secondary: CTR, CPC.
- Install tracking: not available yet; website has no signup. We will not optimize to conversions for now.
- Current site languages available: English, Spanish, French, German, Italian, Portuguese, Japanese.

---

## Store URL Templates (copy & adapt per week)
- Base URLs
  - Apple: https://apps.apple.com/us/app/decorebator/id6749329064
  - Google Play: https://play.google.com/store/apps/details?id=com.lsimaocosta.decorebator

- Apple tracking (simple): append `?ct={country}_{week}`
  - Example token formats: `jp_w1`, `my_w1`, `br_w3`
  - Final Apple example (Japan, Week 1):
    - https://apps.apple.com/us/app/decorebator/id6749329064?ct=jp_w1
  - Final Apple example (Malaysia, Week 1):
    - https://apps.apple.com/us/app/decorebator/id6749329064?ct=my_w1

- Google Play tracking: append URL‑encoded `referrer` with UTM
  - Base UTM: `utm_source=x&utm_medium=paid&utm_campaign={country}_{week}&utm_content={adset}_{creative}`
  - Must URL‑encode the whole referrer value
  - Final Play example (Japan, Week 1, adset=interests, creative=img1):
    - https://play.google.com/store/apps/details?id=com.lsimaocosta.decorebator&referrer=utm_source%3Dx%26utm_medium%3Dpaid%26utm_campaign%3Djp_w1%26utm_content%3Dinterests_img1
  - Final Play example (Malaysia, Week 1, adset=lookalikes, creative=img2):
    - https://play.google.com/store/apps/details?id=com.lsimaocosta.decorebator&referrer=utm_source%3Dx%26utm_medium%3Dpaid%26utm_campaign%3Dmy_w1%26utm_content%3Dlookalikes_img2

Notes
- Keep `{adset}` consistent with your ad group names (e.g., `interests`, `lookalikes`).
- Keep `{creative}` aligned to the image file label (e.g., `img1`, `img2`).
- We will paste the exact variant used in the weekly report for traceability.

---

## Weekly Workflow (repeat each market)
- [ ] Pick market + language for creatives and targetting.
- [ ] Confirm final URLs:
  - [ ] Apple App Store URL + add `ct={country}_{week}` (Apple campaign token).
  - [ ] Google Play URL + add `referrer=utm_source%3Dx%26utm_medium%3Dpaid%26utm_campaign%3D{country}_{week}`.
- [ ] Create campaign in X Ads: Objective = Website Traffic; Budget = R$ 100/day; Schedule = 7 days.
- [ ] Ad groups (2):
  - [ ] Interests/Conversations: language learning, exam prep, study skills, productivity.
  - [ ] Follower look‑alikes: major language‑app and exam‑prep accounts.
- [ ] Creatives (3–4 per ad group): simple image (1200×1200) + local‑language post text.
- [ ] Go live; verify link destinations and click attribution in X Ads.
- [ ] End‑of‑week report: spend, impressions, link clicks, CTR, CPC; save best/worst variants.

Image spec reminder
- Size: 1200×1200 (1:1) or 1920×1080 (16:9). JPG/PNG.
- On‑image text: ≤8 words. Clear CTA text in local language.

Post text template (edit per market)
> Value first sentence. One feature sentence (AI definitions/images/audio + spaced repetition + quizzes). Clear CTA to the store.

---

## Week 1 — Japan (JP)
- Goal: English learning demand with strong study culture; creatives in Japanese; direct to stores.
- Language: Japanese.
- Targeting: Interests = 英語学習, 資格試験, 勉強法; follower look‑alikes = Duolingo/Memrise/Anki/exam prep.
- KPI target (starting): CTR ≥ 1.5%, CPC ≤ R$ 3.00 (adjust after data).
- On‑image headline ideas (Japanese):
  - “語彙を速く覚える”
  - “AI＋クイズで長く記憶”
  - “今すぐ無料で始める”
- Post text starters (Japanese):
  - “AIが定義・音声・画像を作成。間隔反復とクイズで定着。今すぐ無料で始める。”
  - “毎日10分、語彙を伸ばす。AI＋クイズで効率よく学習。”

Creatives (ready to copy)
- Creative JP‑IMG1（General）
  - On‑image: “語彙を速く覚える”
  - Post text: “AIが定義・発音・画像を自動生成。間隔反復と8種類のクイズで語彙を定着。今すぐ無料で始める。”
  - Apple: https://apps.apple.com/us/app/decorebator/id6749329064?ct=jp_w1
  - Google Play: https://play.google.com/store/apps/details?id=com.lsimaocosta.decorebator&referrer=utm_source%3Dx%26utm_medium%3Dpaid%26utm_campaign%3Djp_w1%26utm_content%3Dinterests_img1

- Creative JP‑IMG2（Exam focus）
  - On‑image: “試験対策に強い語彙力”
  - Post text: “英検・TOEIC・IELTSの語彙強化に。AIが定義/音声/画像を作成。間隔反復＋クイズで毎日コツコツ。無料で試す。”
  - Apple: https://apps.apple.com/us/app/decorebator/id6749329064?ct=jp_w1
  - Google Play: https://play.google.com/store/apps/details?id=com.lsimaocosta.decorebator&referrer=utm_source%3Dx%26utm_medium%3Dpaid%26utm_campaign%3Djp_w1%26utm_content%3Dinterests_img2

- Creative JP‑IMG3（Daily habit）
  - On‑image: “毎日10分、続く学習”
  - Post text: “スキマ時間で語彙を増やす。AI＋クイズ、分析で成長が見える。今すぐ無料。”
  - Apple: https://apps.apple.com/us/app/decorebator/id6749329064?ct=jp_w1
  - Google Play: https://play.google.com/store/apps/details?id=com.lsimaocosta.decorebator&referrer=utm_source%3Dx%26utm_medium%3Dpaid%26utm_campaign%3Djp_w1%26utm_content%3Dinterests_img3

---

## Week 2 — Malaysia (MY)
- Goal: Reach Malay speakers likely to learn English; link straight to stores.
- Language: Malay (primary). English acceptable as secondary in post text.
- Targeting: Interests = language learning, exam prep (MUET/IELTS), study skills. Follower look‑alikes = Duolingo/Memrise/Anki/exam creators.
- KPI target (starting): CTR ≥ 1.5%, CPC ≤ R$ 2.50 (adjust after data).
- On‑image headline ideas (Malay):
  - “Belajar Kosa Kata Lebih Pantas”
  - “AI + Kuiz = Ingat Lebih Lama”
  - “Mula Percuma Hari Ini”
- Post text starters (Malay):
  - “Naik taraf kosa kata dengan AI. Kuiz pintar + pengulangan berjarak. Muat turun sekarang.”
  - “Sedia untuk MUET/IELTS? Bina senarai, AI hasilkan definisi, audio & imej.”

---

## Week 3 — Brazil (BR)
- Goal: English for career/university; high mobile usage; Portuguese creatives.
- Language: Portuguese (BR).
- KPI target: CTR ≥ 1.8%, CPC ≤ R$ 2.20.
- On‑image headline ideas (PT‑BR):
  - “Aprenda Vocabulário Mais Rápido”
  - “IA + Quiz = Memorize por Mais Tempo”
  - “Comece Grátis Hoje”
- Post text starters (PT‑BR):
  - “Crie sua lista, a IA gera definições, áudio e imagens. Quiz + repetição espaçada.”
  - “10 min por dia para evoluir seu inglês. Comece grátis.”

---

## Week 4 — Mexico (MX)
- Goal: English for work/exams; Spanish creatives.
- Language: Spanish (LatAm).
- KPI target: CTR ≥ 1.6%, CPC ≤ R$ 2.30.
- On‑image headline ideas (ES):
  - “Aprende vocabulario más rápido”
  - “IA + cuestionarios = recuerda por más tiempo”
  - “Empieza gratis hoy”
- Post text starters (ES):
  - “Crea tu lista; la IA genera definiciones, audio e imágenes. Repetición espaciada + cuestionarios.”
  - “10 min al día para avanzar en tu inglés. Empieza gratis.”

---

## Week 5 — Turkey (TR)
- Goal: English (and some German) learners; price‑sensitive; Turkish creatives.
- Language: Turkish.
- KPI target: CTR ≥ 1.5%, CPC ≤ R$ 2.40.
- On‑image headline ideas (TR):
  - “Kelime Hazneni Daha Hızlı Geliştir”
  - “Yapay Zeka + Quiz = Daha Kalıcı Öğrenme”
  - “Hemen Ücretsiz Başla”
- Post text starters (TR):
  - “Listeyi oluştur, Yapay Zeka tanım, ses ve görsel üretsin. Aralıklı tekrar + quiz.”
  - “Günde 10 dakika ile İngilizcini geliştir. Ücretsiz dene.”

---

## Week 6 — Spain (ES)
- Goal: English learners; polyglots; Spanish creatives for Spain.
- Language: Spanish (EU).
- KPI target: CTR ≥ 1.6%, CPC ≤ R$ 2.60.
- On‑image headline ideas (ES):
  - “Aprende vocabulario más rápido”
  - “IA + quiz = memoria duradera”
  - “Empieza gratis”
- Post text starters (ES):
  - “IA crea definiciones, audio e imágenes. Repetición espaciada + quizzes.”

---

## Week 7 — Italy (IT)
- Goal: English learners; strong acceptance of spaced repetition; Italian creatives.
- Language: Italian.
- KPI target: CTR ≥ 1.6%, CPC ≤ R$ 2.60.
- On‑image headline ideas (IT):
  - “Impara vocabolario più velocemente”
  - “IA + quiz = ricordi più a lungo”
  - “Inizia gratis oggi”
- Post text starters (IT):
  - “Crea la lista; l’IA genera definizioni, audio e immagini. Ripetizione dilazionata + quiz.”

---

## Week 8 — France (FR)
- Goal: English learners; quality/education oriented; French creatives.
- Language: French.
- KPI target: CTR ≥ 1.5%, CPC ≤ R$ 2.80.
- On‑image headline ideas (FR):
  - “Apprenez du vocabulaire plus vite”
  - “IA + quiz = mémorisation durable”
  - “Commencez gratuitement”
- Post text starters (FR):
  - “Créez votre liste; l’IA génère définitions, audio et images. Répétition espacée + quiz.”

---

## Week 9 — Germany (DE)
- Goal: English learners; structured learning appeal; German creatives.
- Language: German.
- KPI target: CTR ≥ 1.5%, CPC ≤ R$ 3.00.
- On‑image headline ideas (DE):
  - “Lerne Vokabeln schneller”
  - “KI + Quiz = länger merken”
  - “Jetzt kostenlos starten”
- Post text starters (DE):
- “Erstelle deine Liste; KI erzeugt Definitionen, Audio und Bilder. Spaced Repetition + Quiz.”

---

## Week 10 — Portugal (PT)
- Goal: English for career/uni; Portuguese (EU) creatives.
- Language: Portuguese (EU).
- KPI target: CTR ≥ 1.6%, CPC ≤ R$ 2.60.
- On‑image headline ideas (PT‑EU):
  - “Aprende vocabulário mais rápido”
  - “IA + quiz = memória duradoura”
  - “Começa grátis”
- Post text starters (PT‑EU):
  - “Cria a tua lista; a IA gera definições, áudio e imagens. Repetição espaçada + quizzes.”

---

## Parking Lot (save for later waves)
- South Korea (KR) — local creatives; link to stores directly.
- Egypt (EG) / Saudi expansion (AR) — Arabic creatives; direct‑to‑store links.
- Thailand (TH), Indonesia (ID), Vietnam (VN), Philippines (PH) — local creatives; consider light EN fallback.

---

## Minimal Reporting Template (fill weekly)
- Market & week: ____ / W__
- Budget (7 days): R$ 700 (R$ 100/day)
- Spend: R$ ____
- Impressions: ____
- Link clicks (to stores): ____
- CTR: ____ %
- CPC: R$ ____
- Best creative (image filename + text): ____
- Notes / next actions: ____

---

## Open Questions for You
- Please paste the final Apple and Google Play URLs so I can apply `ct` (Apple) and `referrer` (Google) parameters consistently.
- Confirm the desired market order. I proposed MY → JP → BR → MX → TR → ES → IT → FR → DE → PT. Want to swap any?
- Time zone for daily pacing and reports?
- Do you want me to generate the first 3 image assets (1200×1200) for Week 1 (Malay) now?

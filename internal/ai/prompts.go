package ai

import (
	"fmt"

	"github.com/rawdah/rawdah-api/internal/models"
)

// BuildHadithPrompt creates a self-contained quiz generation prompt.
// The AI selects an authentic hadith and generates questions — no pre-fetched hadith required.
func BuildHadithPrompt(childAge int, difficulty string) string {
	ageContext := ""
	if childAge > 0 {
		ageContext = fmt.Sprintf(" The child is %d years old, so calibrate language and difficulty appropriately.", childAge)
	}

	return fmt.Sprintf(`You are an Islamic educator creating quiz questions for Muslim children.
Respond ONLY with valid JSON. No preamble, no explanation outside JSON.%s

Choose one authentic hadith from the Kutub al-Sittah (Sahih Bukhari, Sahih Muslim, Sunan Abu Dawud, Jami at-Tirmidhi, Sunan an-Nasai, or Sunan Ibn Majah) graded sahih or hasan, suitable for a child at difficulty level "%s".

IMPORTANT RULES:
- Only use hadiths with a verified, established chain of narration (sahih or hasan). Never fabricate or paraphrase beyond the meaning.
- The text_en must be a faithful English translation of the actual hadith.
- The source must name the specific collection (e.g. "Bukhari", "Muslim", "Abu Dawud").
- If the hadith has a well-known Arabic text, include it in text_ar; otherwise leave text_ar as an empty string.

Then generate exactly 3 multiple choice questions that test comprehension and memorisation of that hadith. Keep language simple and age-appropriate. Each question must have exactly 4 options labeled A, B, C, D.

Respond with exactly this JSON structure:
{"hadith":{"text_en":"...","text_ar":"...","source":"Bukhari","topic":"..."},"questions":[{"id":"q1","question":"...","options":{"A":"...","B":"...","C":"...","D":"..."},"correct_answer":"A","explanation":"..."}]}`,
		ageContext,
		difficulty,
	)
}

// BuildProphetPrompt creates a quiz generation prompt for a prophet's story.
func BuildProphetPrompt(prophet models.Prophet, childAge int) string {
	ageContext := ""
	if childAge > 0 {
		ageContext = fmt.Sprintf(" The child is %d years old, so calibrate language and difficulty appropriately.", childAge)
	}

	nameAr := ""
	if prophet.NameAr != nil {
		nameAr = fmt.Sprintf(" (%s)", *prophet.NameAr)
	}

	miracles := ""
	if prophet.KeyMiracles != nil {
		miracles = fmt.Sprintf("\nKey Miracles: %s", *prophet.KeyMiracles)
	}

	nation := ""
	if prophet.Nation != nil {
		nation = fmt.Sprintf("\nNation: %s", *prophet.Nation)
	}

	quranRefs := ""
	if prophet.QuranRefs != nil {
		quranRefs = fmt.Sprintf("\nQuran References: %s", *prophet.QuranRefs)
	}

	return fmt.Sprintf(`You are an Islamic educator. Respond ONLY with valid JSON.%s

Generate exactly 3 multiple choice questions for a child about Prophet %s%s (peace be upon him).

Context:
- Story: %s%s%s%s
- Difficulty: %s

Mix question types: factual recall, comprehension, and fill-in-the-blank.
Keep language simple and encouraging. Each question must have exactly 4 options labeled A, B, C, D.

Respond with exactly this JSON structure:
{"questions":[{"id":"q1","question":"...","options":{"A":"...","B":"...","C":"...","D":"..."},"correct_answer":"A","explanation":"..."}]}`,
		ageContext,
		prophet.NameEn,
		nameAr,
		prophet.StorySummary,
		miracles,
		nation,
		quranRefs,
		prophet.Difficulty,
	)
}

// BuildQuranPrompt creates a quiz generation prompt for a Quran verse.
func BuildQuranPrompt(verse models.QuranVerse, childAge int) string {
	ageContext := ""
	if childAge > 0 {
		ageContext = fmt.Sprintf(" The child is %d years old, so calibrate language and difficulty appropriately.", childAge)
	}

	transliteration := ""
	if verse.Transliteration != nil {
		transliteration = fmt.Sprintf("\nTransliteration: %s", *verse.Transliteration)
	}

	lifeApp := ""
	if verse.LifeApplication != nil {
		lifeApp = fmt.Sprintf("\nLife Application: %s", *verse.LifeApplication)
	}

	topic := ""
	if verse.Topic != nil {
		topic = fmt.Sprintf("\nTopic: %s", *verse.Topic)
	}

	return fmt.Sprintf(`You are an Islamic educator. Respond ONLY with valid JSON.%s

Generate 2 comprehension questions and 1 fill-in-the-blank for a child about this Quran verse:

Surah: %s (%d:%d)
Arabic: %s%s
English: %s
Meaning: %s%s%s
Difficulty: %s

Focus on understanding the meaning, not Arabic memorisation. Keep language simple.
Each question must have exactly 4 options labeled A, B, C, D.

Respond with exactly this JSON structure:
{"questions":[{"id":"q1","question":"...","options":{"A":"...","B":"...","C":"...","D":"..."},"correct_answer":"A","explanation":"..."}]}`,
		ageContext,
		verse.SurahNameEn,
		verse.SurahNumber,
		verse.AyahNumber,
		verse.TextAr,
		transliteration,
		verse.TextEn,
		verse.TafsirSimple,
		lifeApp,
		topic,
		verse.Difficulty,
	)
}

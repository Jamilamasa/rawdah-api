//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type prophet struct {
	NameEn       string
	NameAr       *string
	OrderNum     *int
	StorySummary string
	KeyMiracles  *string
	Nation       string
	QuranRefs    string
	Difficulty   string
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

var prophets = []prophet{
	{
		NameEn:       "Adam",
		NameAr:       strPtr("آدم"),
		OrderNum:     intPtr(1),
		StorySummary: "Adam was the first human being and the first prophet, created by Allah from clay. He and his wife Hawwa (Eve) were placed in Paradise but were sent to Earth after eating from the forbidden tree. Allah taught Adam the names of all things, demonstrating humanity's capacity for knowledge.",
		KeyMiracles:  strPtr("Created directly by Allah from clay without parents; taught the names of all things"),
		Nation:       "Humanity",
		QuranRefs:    "2:30-39, 7:11-25, 20:115-123",
		Difficulty:   "easy",
	},
	{
		NameEn:       "Idris",
		NameAr:       strPtr("إدريس"),
		OrderNum:     intPtr(2),
		StorySummary: "Idris (Enoch) was a righteous prophet known for his wisdom and knowledge. He was the first to use a pen for writing and was skilled in mathematics and astronomy. Allah raised him to a high station.",
		KeyMiracles:  strPtr("Raised to a high station by Allah; first to write with a pen"),
		Nation:       "Babylon",
		QuranRefs:    "19:56-57, 21:85",
		Difficulty:   "medium",
	},
	{
		NameEn:       "Nuh",
		NameAr:       strPtr("نوح"),
		OrderNum:     intPtr(3),
		StorySummary: "Nuh (Noah) preached monotheism to his people for 950 years. When they persistently refused to believe, Allah commanded him to build an ark. A great flood came and destroyed the disbelievers while Nuh, the believers, and pairs of animals were saved.",
		KeyMiracles:  strPtr("Built the Ark under divine guidance; survived the Great Flood"),
		Nation:       "Ancient Mesopotamia",
		QuranRefs:    "7:59-64, 11:25-48, 26:105-122, 71:1-28",
		Difficulty:   "easy",
	},
	{
		NameEn:       "Hud",
		NameAr:       strPtr("هود"),
		OrderNum:     intPtr(4),
		StorySummary: "Hud was sent to the powerful tribe of Aad in ancient Arabia. They were known for their great strength and tall buildings. When they rejected Hud's message, Allah destroyed them with a violent wind that lasted seven nights and eight days.",
		KeyMiracles:  strPtr("Saved from the devastating windstorm that destroyed the Aad"),
		Nation:       "Aad (Arabia)",
		QuranRefs:    "7:65-72, 11:50-60, 26:123-140",
		Difficulty:   "medium",
	},
	{
		NameEn:       "Salih",
		NameAr:       strPtr("صالح"),
		OrderNum:     intPtr(5),
		StorySummary: "Salih was sent to the tribe of Thamud who carved homes in mountains. As a sign, Allah sent them a she-camel and commanded them not to harm her. When they slaughtered the camel, Allah destroyed them.",
		KeyMiracles:  strPtr("The miraculous she-camel that emerged from a rock"),
		Nation:       "Thamud (Arabia)",
		QuranRefs:    "7:73-79, 11:61-68, 26:141-159",
		Difficulty:   "medium",
	},
	{
		NameEn:       "Ibrahim",
		NameAr:       strPtr("إبراهيم"),
		OrderNum:     intPtr(6),
		StorySummary: "Ibrahim (Abraham) is known as the Friend of Allah (Khalilullah). He destroyed the idols of his people and was thrown into a massive fire but was saved by Allah. He built the Kaaba with his son Ismail and was tested to sacrifice his son before Allah replaced the sacrifice.",
		KeyMiracles:  strPtr("Survived being thrown into fire; built the Kaaba; birds brought back to life"),
		Nation:       "Babylon/Canaan",
		QuranRefs:    "2:124-141, 6:74-83, 19:41-50, 21:51-73",
		Difficulty:   "easy",
	},
	{
		NameEn:       "Lut",
		NameAr:       strPtr("لوط"),
		OrderNum:     intPtr(7),
		StorySummary: "Lut (Lot) was Ibrahim's nephew and was sent to the people of Sodom and Gomorrah. He preached against their immoral practices. When they refused to reform, Allah destroyed their cities and saved Lut and his family, except his wife.",
		KeyMiracles:  strPtr("Saved from the destruction of Sodom with his believing family"),
		Nation:       "Sodom and Gomorrah",
		QuranRefs:    "7:80-84, 11:77-83, 26:160-175",
		Difficulty:   "medium",
	},
	{
		NameEn:       "Ismail",
		NameAr:       strPtr("إسماعيل"),
		OrderNum:     intPtr(8),
		StorySummary: "Ismail (Ishmael) was the eldest son of Ibrahim. As an infant, he and his mother Hajar were left in the desert of Makkah where the spring of Zamzam emerged. He helped his father build the Kaaba and was known for his patience when his father was commanded to sacrifice him.",
		KeyMiracles:  strPtr("Zamzam spring; survived the command of sacrifice"),
		Nation:       "Arabia",
		QuranRefs:    "2:125-127, 19:54-55, 37:101-109",
		Difficulty:   "easy",
	},
	{
		NameEn:       "Ishaq",
		NameAr:       strPtr("إسحاق"),
		OrderNum:     intPtr(9),
		StorySummary: "Ishaq (Isaac) was the second son of Ibrahim, born when his parents were very old as a miracle from Allah. He continued his father's mission of monotheism and became a prophet sent to the people of Canaan.",
		KeyMiracles:  strPtr("Miraculous birth to elderly parents"),
		Nation:       "Canaan",
		QuranRefs:    "11:71-73, 19:49, 37:112-113",
		Difficulty:   "easy",
	},
	{
		NameEn:       "Yaqub",
		NameAr:       strPtr("يعقوب"),
		OrderNum:     intPtr(10),
		StorySummary: "Yaqub (Jacob), also called Israel, was the son of Ishaq. He is the father of the twelve tribes of Israel. He was deeply patient through many trials including the separation from his beloved son Yusuf.",
		KeyMiracles:  strPtr("His sight was restored when Yusuf's shirt was placed on his face"),
		Nation:       "Canaan",
		QuranRefs:    "12:4-101, 19:49, 21:72-73",
		Difficulty:   "medium",
	},
	{
		NameEn:       "Yusuf",
		NameAr:       strPtr("يوسف"),
		OrderNum:     intPtr(11),
		StorySummary: "Yusuf (Joseph) was thrown into a well by his brothers out of jealousy. He was sold into slavery in Egypt, falsely imprisoned, but eventually became a senior official in Egypt. His story is called the best of stories in the Quran.",
		KeyMiracles:  strPtr("Ability to interpret dreams; surviving the well and prison"),
		Nation:       "Egypt",
		QuranRefs:    "12:1-111",
		Difficulty:   "easy",
	},
	{
		NameEn:       "Shuaib",
		NameAr:       strPtr("شعيب"),
		OrderNum:     intPtr(12),
		StorySummary: "Shuaib was sent to the people of Madyan who were known for cheating in trade and transactions. He commanded them to be honest in their dealings. When they refused, Allah destroyed them with an earthquake and thunderbolt.",
		KeyMiracles:  strPtr("Saved from the earthquake and thunderbolt that destroyed Madyan"),
		Nation:       "Madyan",
		QuranRefs:    "7:85-93, 11:84-95, 26:176-191",
		Difficulty:   "medium",
	},
	{
		NameEn:       "Ayyub",
		NameAr:       strPtr("أيوب"),
		OrderNum:     intPtr(13),
		StorySummary: "Ayyub (Job) is the prophet of patience. He was tested with severe illness, loss of wealth and family for many years. Despite extreme suffering, he remained patient and thankful to Allah. Allah eventually restored his health and doubled his blessings.",
		KeyMiracles:  strPtr("Miraculous recovery from prolonged illness after years of patient supplication"),
		Nation:       "Arabia or Jordan",
		QuranRefs:    "21:83-84, 38:41-44",
		Difficulty:   "easy",
	},
	{
		NameEn:       "Musa",
		NameAr:       strPtr("موسى"),
		OrderNum:     intPtr(14),
		StorySummary: "Musa (Moses) is one of the most mentioned prophets in the Quran. Born during Pharaoh's massacre of infant boys, he was raised in Pharaoh's palace. Allah spoke to him directly and gave him many miracles. He led the Israelites out of Egypt and received the Torah.",
		KeyMiracles:  strPtr("Staff turned into serpent; hand glowing white; parting the Red Sea; receiving the Torah"),
		Nation:       "Israelites (Egypt)",
		QuranRefs:    "2:49-61, 7:103-162, 20:9-98, 26:10-68, 28:1-46",
		Difficulty:   "easy",
	},
	{
		NameEn:       "Harun",
		NameAr:       strPtr("هارون"),
		OrderNum:     intPtr(15),
		StorySummary: "Harun (Aaron) was the brother of Musa and his helper. He was eloquent and helped Musa convey the message to Pharaoh. He was appointed as prophet alongside Musa and was left in charge of the Israelites while Musa went to receive the Torah.",
		KeyMiracles:  strPtr("Shared in the miracles of Musa"),
		Nation:       "Israelites (Egypt)",
		QuranRefs:    "7:142, 20:29-36, 25:35",
		Difficulty:   "medium",
	},
	{
		NameEn:       "Dhul-Kifl",
		NameAr:       strPtr("ذو الكفل"),
		OrderNum:     intPtr(16),
		StorySummary: "Dhul-Kifl is mentioned among the patient and righteous prophets. He is believed to have been given a great responsibility to guide his people and was known for his patience and steadfastness in worship.",
		Nation:       "Believed to be in Syria or Iraq",
		QuranRefs:    "21:85-86, 38:48",
		Difficulty:   "hard",
	},
	{
		NameEn:       "Dawud",
		NameAr:       strPtr("داود"),
		OrderNum:     intPtr(17),
		StorySummary: "Dawud (David) was a great prophet and king. As a young man, he killed the giant Jalut (Goliath) with a sling. Allah gave him wisdom, prophethood, and kingship. The Psalms (Zabur) were revealed to him. Iron became soft in his hands.",
		KeyMiracles:  strPtr("Iron softened in his hands; birds and mountains praised Allah with him; the Zabur revealed to him"),
		Nation:       "Israelites (Jerusalem)",
		QuranRefs:    "2:251, 6:84, 21:78-80, 34:10-11, 38:17-26",
		Difficulty:   "easy",
	},
	{
		NameEn:       "Sulayman",
		NameAr:       strPtr("سليمان"),
		OrderNum:     intPtr(18),
		StorySummary: "Sulayman (Solomon) was the son of Dawud and a mighty prophet and king. Allah granted him control over winds, jinns, and the ability to understand the speech of birds and animals. He ruled a great kingdom and built the great temple in Jerusalem.",
		KeyMiracles:  strPtr("Control over wind, jinn, and animals; understanding the language of birds and animals; a vast kingdom"),
		Nation:       "Israelites (Jerusalem)",
		QuranRefs:    "21:81-82, 27:15-44, 34:12-14, 38:30-40",
		Difficulty:   "easy",
	},
	{
		NameEn:       "Ilyas",
		NameAr:       strPtr("إلياس"),
		OrderNum:     intPtr(19),
		StorySummary: "Ilyas (Elijah) was sent to the people of Baal in the land of Canaan who worshipped an idol called Baal. He called them to the worship of Allah alone but most rejected him. He is praised as one of the righteous in the Quran.",
		Nation:       "Canaan",
		QuranRefs:    "6:85, 37:123-132",
		Difficulty:   "hard",
	},
	{
		NameEn:       "Al-Yasa",
		NameAr:       strPtr("اليسع"),
		OrderNum:     intPtr(20),
		StorySummary: "Al-Yasa (Elisha) was the student and successor of Ilyas. He continued the mission of Ilyas among the Israelites and is praised alongside other great prophets in the Quran.",
		Nation:       "Israelites (Canaan)",
		QuranRefs:    "6:86, 38:48",
		Difficulty:   "hard",
	},
	{
		NameEn:       "Yunus",
		NameAr:       strPtr("يونس"),
		OrderNum:     intPtr(21),
		StorySummary: "Yunus (Jonah) was sent to the people of Nineveh. When they rejected him, he left them and was swallowed by a large whale. In the belly of the whale, he called upon Allah in sincere repentance with the prayer Laa ilaaha illa anta subhaanaka inni kuntu minal-dhaalimeen. Allah saved him and his people eventually believed.",
		KeyMiracles:  strPtr("Survived inside the belly of a whale through divine mercy"),
		Nation:       "Nineveh (Iraq)",
		QuranRefs:    "10:98, 21:87-88, 37:139-148, 68:48-50",
		Difficulty:   "easy",
	},
	{
		NameEn:       "Zakariyya",
		NameAr:       strPtr("زكريا"),
		OrderNum:     intPtr(22),
		StorySummary: "Zakariyya (Zechariah) was a devoted prophet who took care of Maryam (Mary). Despite being old and his wife being barren, he prayed to Allah for a child. Allah granted him the miraculous birth of Yahya (John).",
		KeyMiracles:  strPtr("Son Yahya born miraculously to an elderly barren couple"),
		Nation:       "Israelites (Jerusalem)",
		QuranRefs:    "3:37-41, 6:85, 19:2-11, 21:89-90",
		Difficulty:   "medium",
	},
	{
		NameEn:       "Yahya",
		NameAr:       strPtr("يحيى"),
		OrderNum:     intPtr(23),
		StorySummary: "Yahya (John the Baptist) was born as a miracle to aged parents Zakariyya and his barren wife. He was given wisdom while still a child and was known for his piety, chastity, and devotion to Allah. He confirmed the coming of Isa (Jesus).",
		KeyMiracles:  strPtr("Miraculous birth; granted wisdom as a child"),
		Nation:       "Israelites (Jerusalem)",
		QuranRefs:    "3:39, 6:85, 19:12-15, 21:90",
		Difficulty:   "medium",
	},
	{
		NameEn:       "Isa",
		NameAr:       strPtr("عيسى"),
		OrderNum:     intPtr(24),
		StorySummary: "Isa (Jesus) was born miraculously to the virgin Maryam without a father. He spoke as a baby in the cradle, healed the blind and lepers, and brought the dead back to life — all with Allah's permission. The Injeel (Gospel) was revealed to him. He was raised to heaven and will return before the Day of Judgment.",
		KeyMiracles:  strPtr("Born without a father; spoke in the cradle; healed the blind and lepers; brought the dead back to life"),
		Nation:       "Israelites (Palestine)",
		QuranRefs:    "3:45-55, 4:157-159, 5:110-120, 19:16-37",
		Difficulty:   "easy",
	},
	{
		NameEn:       "Muhammad",
		NameAr:       strPtr("محمد"),
		OrderNum:     intPtr(25),
		StorySummary: "Muhammad ﷺ is the final messenger and prophet, the Seal of the Prophets. Born in Makkah around 570 CE, he received the first revelation in the Cave of Hira at age 40. He endured 13 years of persecution in Makkah before migrating to Madinah. He united the Arabian Peninsula under Islam. The Quran was revealed to him as the final and preserved word of Allah.",
		KeyMiracles:  strPtr("The Quran — the eternal miracle; Isra and Miraj (Night Journey); splitting of the moon; water flowing from his fingers"),
		Nation:       "Arabia",
		QuranRefs:    "33:40, 47:2, 48:29, 33:56",
		Difficulty:   "easy",
	},
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM prophets`).Scan(&count); err != nil {
		log.Fatalf("count prophets: %v", err)
	}
	if count > 0 {
		fmt.Printf("prophets table already has %d rows, skipping seed\n", count)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalf("begin tx: %v", err)
	}

	inserted := 0
	for _, p := range prophets {
		var quranRefs *string
		if p.QuranRefs != "" {
			quranRefs = strPtr(p.QuranRefs)
		}
		var nation *string
		if p.Nation != "" {
			nation = strPtr(p.Nation)
		}
		var keyMiracles *string
		if p.KeyMiracles != nil {
			keyMiracles = p.KeyMiracles
		}

		_, err := tx.ExecContext(ctx,
			`INSERT INTO prophets (name_en, name_ar, order_num, story_summary, key_miracles, nation, quran_refs, difficulty)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT (order_num) DO NOTHING`,
			p.NameEn, p.NameAr, p.OrderNum, p.StorySummary, keyMiracles, nation, quranRefs, p.Difficulty,
		)
		if err != nil {
			tx.Rollback()
			log.Fatalf("insert prophet '%s': %v", p.NameEn, err)
		}
		inserted++
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("commit: %v", err)
	}

	fmt.Printf("seeded %d prophets successfully\n", inserted)
}

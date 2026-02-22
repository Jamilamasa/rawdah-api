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

type quranVerse struct {
	SurahNumber     int
	AyahNumber      int
	SurahNameEn     string
	TextAr          string
	TextEn          string
	Transliteration *string
	TafsirSimple    string
	LifeApplication *string
	Topic           *string
	Difficulty      string
}

func ptrStr(s string) *string { return &s }

var verses = []quranVerse{
	// --- Al-Fatiha (1) ---
	{
		SurahNumber:     1,
		AyahNumber:      1,
		SurahNameEn:     "Al-Fatiha",
		TextAr:          "بِسْمِ اللَّهِ الرَّحْمَٰنِ الرَّحِيمِ",
		TextEn:          "In the name of Allah, the Most Gracious, the Most Merciful.",
		Transliteration: ptrStr("Bismillahi r-rahmani r-rahim"),
		TafsirSimple:    "We begin everything with the name of Allah, reminding ourselves that He is full of mercy and love for us.",
		LifeApplication: ptrStr("Say Bismillah before starting any good action."),
		Topic:           ptrStr("tawheed"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     1,
		AyahNumber:      2,
		SurahNameEn:     "Al-Fatiha",
		TextAr:          "الْحَمْدُ لِلَّهِ رَبِّ الْعَالَمِينَ",
		TextEn:          "All praise is due to Allah, Lord of all the worlds.",
		Transliteration: ptrStr("Alhamdu lillahi rabbi l-alamin"),
		TafsirSimple:    "All praise and gratitude belong only to Allah, who created and takes care of everything in existence.",
		LifeApplication: ptrStr("Say Alhamdulillah whenever something good happens."),
		Topic:           ptrStr("gratitude"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     1,
		AyahNumber:      3,
		SurahNameEn:     "Al-Fatiha",
		TextAr:          "الرَّحْمَٰنِ الرَّحِيمِ",
		TextEn:          "The Most Gracious, the Most Merciful.",
		Transliteration: ptrStr("Ar-rahmani r-rahim"),
		TafsirSimple:    "Two of Allah's most beautiful names: Ar-Rahman (the Infinitely Merciful to all) and Ar-Raheem (the especially Merciful to believers).",
		Topic:           ptrStr("tawheed"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     1,
		AyahNumber:      4,
		SurahNameEn:     "Al-Fatiha",
		TextAr:          "مَالِكِ يَوْمِ الدِّينِ",
		TextEn:          "Master of the Day of Judgment.",
		Transliteration: ptrStr("Maliki yawmi d-din"),
		TafsirSimple:    "Allah is the Master and Judge on the Day when everyone will be held accountable for their deeds.",
		LifeApplication: ptrStr("Remember that Allah watches our actions, so always try to do what is right."),
		Topic:           ptrStr("akhirah"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     1,
		AyahNumber:      5,
		SurahNameEn:     "Al-Fatiha",
		TextAr:          "إِيَّاكَ نَعْبُدُ وَإِيَّاكَ نَسْتَعِينُ",
		TextEn:          "You alone we worship, and You alone we ask for help.",
		Transliteration: ptrStr("Iyyaka na'budu wa iyyaka nasta'in"),
		TafsirSimple:    "We direct all our worship only to Allah and we only ask Him for help. He is the only one truly worthy of worship.",
		LifeApplication: ptrStr("When you need help, make du'a to Allah before anything else."),
		Topic:           ptrStr("tawheed"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     1,
		AyahNumber:      6,
		SurahNameEn:     "Al-Fatiha",
		TextAr:          "اهْدِنَا الصِّرَاطَ الْمُسْتَقِيمَ",
		TextEn:          "Guide us to the straight path.",
		Transliteration: ptrStr("Ihdina s-sirata l-mustaqim"),
		TafsirSimple:    "We ask Allah to keep us on the right path — the path of Islam, of doing good and avoiding evil.",
		LifeApplication: ptrStr("Ask Allah for guidance in your decisions and choices every day."),
		Topic:           ptrStr("guidance"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     1,
		AyahNumber:      7,
		SurahNameEn:     "Al-Fatiha",
		TextAr:          "صِرَاطَ الَّذِينَ أَنْعَمْتَ عَلَيْهِمْ غَيْرِ الْمَغْضُوبِ عَلَيْهِمْ وَلَا الضَّالِّينَ",
		TextEn:          "The path of those upon whom You have bestowed favor, not of those who have evoked anger or of those who are astray.",
		Transliteration: ptrStr("Sirata alladhina an'amta alayhim ghayri l-maghdubi alayhim wa la d-dallin"),
		TafsirSimple:    "We ask to follow the path of the prophets and righteous believers, not the path of those who disobeyed Allah or went astray.",
		Topic:           ptrStr("guidance"),
		Difficulty:      "medium",
	},
	// --- Ayat al-Kursi (2:255) ---
	{
		SurahNumber:     2,
		AyahNumber:      255,
		SurahNameEn:     "Al-Baqarah",
		TextAr:          "اللَّهُ لَا إِلَٰهَ إِلَّا هُوَ الْحَيُّ الْقَيُّومُ ۚ لَا تَأْخُذُهُ سِنَةٌ وَلَا نَوْمٌ",
		TextEn:          "Allah — there is no deity except Him, the Ever-Living, the Sustainer of existence. Neither drowsiness overtakes Him nor sleep.",
		Transliteration: ptrStr("Allahu la ilaha illa huwa l-hayyu l-qayyum, la ta'khudhuhu sinatun wa la nawm"),
		TafsirSimple:    "The greatest verse in the Quran. It tells us that Allah is the only true God, always living and never sleeping. He takes care of all things.",
		LifeApplication: ptrStr("Recite Ayat al-Kursi after every prayer and before sleeping for protection."),
		Topic:           ptrStr("tawheed"),
		Difficulty:      "medium",
	},
	// --- Al-Baqarah 2:286 ---
	{
		SurahNumber:     2,
		AyahNumber:      286,
		SurahNameEn:     "Al-Baqarah",
		TextAr:          "لَا يُكَلِّفُ اللَّهُ نَفْسًا إِلَّا وُسْعَهَا",
		TextEn:          "Allah does not burden a soul beyond that it can bear.",
		Transliteration: ptrStr("La yukallifu llahu nafsan illa wus'aha"),
		TafsirSimple:    "Allah is so merciful that He never gives us more than we can handle. Every difficulty we face is within our ability to bear.",
		LifeApplication: ptrStr("When life feels hard, remember Allah knows you can get through it. Ask Him for strength."),
		Topic:           ptrStr("patience"),
		Difficulty:      "easy",
	},
	// --- Al-Imran 3:190 ---
	{
		SurahNumber:     3,
		AyahNumber:      190,
		SurahNameEn:     "Al-Imran",
		TextAr:          "إِنَّ فِي خَلْقِ السَّمَاوَاتِ وَالْأَرْضِ وَاخْتِلَافِ اللَّيْلِ وَالنَّهَارِ لَآيَاتٍ لِأُولِي الْأَلْبَابِ",
		TextEn:          "Indeed, in the creation of the heavens and the earth and the alternation of the night and the day are signs for those of understanding.",
		Transliteration: ptrStr("Inna fi khalqi s-samawati wa l-ardi wa khtilafi l-layli wa n-nahari la'ayatin li'uli l-albab"),
		TafsirSimple:    "Looking at the sky, the earth, and how day turns to night shows us the greatness of Allah. These are signs for people who think and reflect.",
		LifeApplication: ptrStr("Look at nature around you and think about how amazing Allah's creation is."),
		Topic:           ptrStr("creation"),
		Difficulty:      "medium",
	},
	// --- Al-Asr (103) ---
	{
		SurahNumber:     103,
		AyahNumber:      1,
		SurahNameEn:     "Al-Asr",
		TextAr:          "وَالْعَصْرِ",
		TextEn:          "By time,",
		Transliteration: ptrStr("Wal-asr"),
		TafsirSimple:    "Allah swears by time, showing how precious and valuable it is. We should not waste our time.",
		Topic:           ptrStr("wisdom"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     103,
		AyahNumber:      2,
		SurahNameEn:     "Al-Asr",
		TextAr:          "إِنَّ الْإِنسَانَ لَفِي خُسْرٍ",
		TextEn:          "Indeed, mankind is in loss,",
		Transliteration: ptrStr("Inna l-insana la-fi khusr"),
		TafsirSimple:    "Every human being is losing their time unless they use it wisely for good things. Time is the most valuable thing we have.",
		Topic:           ptrStr("wisdom"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     103,
		AyahNumber:      3,
		SurahNameEn:     "Al-Asr",
		TextAr:          "إِلَّا الَّذِينَ آمَنُوا وَعَمِلُوا الصَّالِحَاتِ وَتَوَاصَوْا بِالْحَقِّ وَتَوَاصَوْا بِالصَّبْرِ",
		TextEn:          "Except for those who have believed and done righteous deeds and advised each other to truth and advised each other to patience.",
		Transliteration: ptrStr("Illa alladhina amanu wa amilu s-salihati wa tawassaw bi-l-haqqi wa tawassaw bi-s-sabr"),
		TafsirSimple:    "Only those who believe in Allah, do good deeds, remind each other of the truth, and encourage each other to be patient will be truly successful.",
		LifeApplication: ptrStr("Be a good friend who encourages others to do what is right and to be patient."),
		Topic:           ptrStr("wisdom"),
		Difficulty:      "medium",
	},
	// --- Al-Kawthar (108) ---
	{
		SurahNumber:     108,
		AyahNumber:      1,
		SurahNameEn:     "Al-Kawthar",
		TextAr:          "إِنَّا أَعْطَيْنَاكَ الْكَوْثَرَ",
		TextEn:          "Indeed, We have granted you, [O Muhammad], al-Kawthar.",
		Transliteration: ptrStr("Inna a'taynaka l-kawthar"),
		TafsirSimple:    "Allah granted Prophet Muhammad a river in Paradise called Al-Kawthar and abundant blessings in this life and the next.",
		Topic:           ptrStr("prophet"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     108,
		AyahNumber:      2,
		SurahNameEn:     "Al-Kawthar",
		TextAr:          "فَصَلِّ لِرَبِّكَ وَانْحَرْ",
		TextEn:          "So pray to your Lord and sacrifice [to Him alone].",
		Transliteration: ptrStr("Fasalli li-rabbika wa-nhar"),
		TafsirSimple:    "Allah commands us to pray and offer sacrifice only for Him, not for anyone or anything else.",
		LifeApplication: ptrStr("Perform your prayers with full dedication for Allah alone."),
		Topic:           ptrStr("prayer"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     108,
		AyahNumber:      3,
		SurahNameEn:     "Al-Kawthar",
		TextAr:          "إِنَّ شَانِئَكَ هُوَ الْأَبْتَرُ",
		TextEn:          "Indeed, your enemy is the one cut off.",
		Transliteration: ptrStr("Inna shani'aka huwa l-abtar"),
		TafsirSimple:    "Those who hate the Prophet and Islam will be the ones forgotten and cut off, while the Prophet's legacy and Islam will continue forever.",
		Topic:           ptrStr("prophet"),
		Difficulty:      "medium",
	},
	// --- Al-Ikhlas (112) ---
	{
		SurahNumber:     112,
		AyahNumber:      1,
		SurahNameEn:     "Al-Ikhlas",
		TextAr:          "قُلْ هُوَ اللَّهُ أَحَدٌ",
		TextEn:          "Say: He is Allah, the One.",
		Transliteration: ptrStr("Qul huwa llahu ahad"),
		TafsirSimple:    "Allah commands us to declare His oneness. Allah is one and completely unique — there is nothing and no one like Him.",
		LifeApplication: ptrStr("Recite Surah Al-Ikhlas three times — it equals reciting one-third of the Quran in reward."),
		Topic:           ptrStr("tawheed"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     112,
		AyahNumber:      2,
		SurahNameEn:     "Al-Ikhlas",
		TextAr:          "اللَّهُ الصَّمَدُ",
		TextEn:          "Allah, the Eternal Refuge.",
		Transliteration: ptrStr("Allahu s-samad"),
		TafsirSimple:    "As-Samad means Allah is the one everyone depends on and turns to in times of need. He depends on no one but everything depends on Him.",
		Topic:           ptrStr("tawheed"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     112,
		AyahNumber:      3,
		SurahNameEn:     "Al-Ikhlas",
		TextAr:          "لَمْ يَلِدْ وَلَمْ يُولَدْ",
		TextEn:          "He neither begets nor is born,",
		Transliteration: ptrStr("Lam yalid wa lam yulad"),
		TafsirSimple:    "Allah has no children and no parents. He has always existed and will always exist. He is completely unique unlike any created being.",
		Topic:           ptrStr("tawheed"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     112,
		AyahNumber:      4,
		SurahNameEn:     "Al-Ikhlas",
		TextAr:          "وَلَمْ يَكُن لَّهُ كُفُوًا أَحَدٌ",
		TextEn:          "Nor is there to Him any equivalent.",
		Transliteration: ptrStr("Wa lam yakun lahu kufuwan ahad"),
		TafsirSimple:    "There is absolutely nothing equal or similar to Allah. He is beyond all comparison and description.",
		Topic:           ptrStr("tawheed"),
		Difficulty:      "easy",
	},
	// --- Al-Falaq (113) ---
	{
		SurahNumber:     113,
		AyahNumber:      1,
		SurahNameEn:     "Al-Falaq",
		TextAr:          "قُلْ أَعُوذُ بِرَبِّ الْفَلَقِ",
		TextEn:          "Say: I seek refuge in the Lord of daybreak,",
		Transliteration: ptrStr("Qul a'udhu bi-rabbi l-falaq"),
		TafsirSimple:    "We seek protection from Allah, the Lord of the dawn, from all things that may harm us.",
		LifeApplication: ptrStr("Recite Al-Falaq morning and evening for protection throughout the day."),
		Topic:           ptrStr("protection"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     113,
		AyahNumber:      2,
		SurahNameEn:     "Al-Falaq",
		TextAr:          "مِن شَرِّ مَا خَلَقَ",
		TextEn:          "From the evil of that which He created,",
		Transliteration: ptrStr("Min sharri ma khalaq"),
		TafsirSimple:    "We ask Allah to protect us from the evil found in His creation, including dangerous animals, diseases, and harmful things.",
		Topic:           ptrStr("protection"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     113,
		AyahNumber:      3,
		SurahNameEn:     "Al-Falaq",
		TextAr:          "وَمِن شَرِّ غَاسِقٍ إِذَا وَقَبَ",
		TextEn:          "And from the evil of darkness when it settles,",
		Transliteration: ptrStr("Wa min sharri ghasiqin idha waqab"),
		TafsirSimple:    "We seek protection from the dangers that come during the night, when harmful things may become more active.",
		Topic:           ptrStr("protection"),
		Difficulty:      "medium",
	},
	{
		SurahNumber:     113,
		AyahNumber:      4,
		SurahNameEn:     "Al-Falaq",
		TextAr:          "وَمِن شَرِّ النَّفَّاثَاتِ فِي الْعُقَدِ",
		TextEn:          "And from the evil of the blowers in knots,",
		Transliteration: ptrStr("Wa min sharri n-naffathati fi l-uqad"),
		TafsirSimple:    "We seek protection from those who practice magic and try to harm others through it. Only Allah can protect us from such things.",
		Topic:           ptrStr("protection"),
		Difficulty:      "hard",
	},
	{
		SurahNumber:     113,
		AyahNumber:      5,
		SurahNameEn:     "Al-Falaq",
		TextAr:          "وَمِن شَرِّ حَاسِدٍ إِذَا حَسَدَ",
		TextEn:          "And from the evil of an envier when he envies.",
		Transliteration: ptrStr("Wa min sharri hasidin idha hasad"),
		TafsirSimple:    "We seek protection from people who are jealous and may wish harm upon us because of their envy.",
		LifeApplication: ptrStr("Never be jealous of others. Be happy when Allah blesses them, and ask Allah for your own blessings."),
		Topic:           ptrStr("protection"),
		Difficulty:      "easy",
	},
	// --- An-Nas (114) ---
	{
		SurahNumber:     114,
		AyahNumber:      1,
		SurahNameEn:     "An-Nas",
		TextAr:          "قُلْ أَعُوذُ بِرَبِّ النَّاسِ",
		TextEn:          "Say: I seek refuge in the Lord of mankind,",
		Transliteration: ptrStr("Qul a'udhu bi-rabbi n-nas"),
		TafsirSimple:    "We seek protection in Allah, who is the Lord and Master of all human beings everywhere.",
		LifeApplication: ptrStr("Recite An-Nas in the morning, evening, and before sleeping."),
		Topic:           ptrStr("protection"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     114,
		AyahNumber:      2,
		SurahNameEn:     "An-Nas",
		TextAr:          "مَلِكِ النَّاسِ",
		TextEn:          "The King of mankind,",
		Transliteration: ptrStr("Maliki n-nas"),
		TafsirSimple:    "Allah is the true King of all people. All earthly kings and rulers are under His authority.",
		Topic:           ptrStr("tawheed"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     114,
		AyahNumber:      3,
		SurahNameEn:     "An-Nas",
		TextAr:          "إِلَٰهِ النَّاسِ",
		TextEn:          "The God of mankind,",
		Transliteration: ptrStr("Ilahi n-nas"),
		TafsirSimple:    "Allah is the only true God of all people, whether they believe in Him or not. He alone deserves to be worshipped.",
		Topic:           ptrStr("tawheed"),
		Difficulty:      "easy",
	},
	{
		SurahNumber:     114,
		AyahNumber:      4,
		SurahNameEn:     "An-Nas",
		TextAr:          "مِن شَرِّ الْوَسْوَاسِ الْخَنَّاسِ",
		TextEn:          "From the evil of the retreating whisperer,",
		Transliteration: ptrStr("Min sharri l-waswasi l-khannas"),
		TafsirSimple:    "We seek protection from Shaytan (Satan), who whispers bad thoughts into our minds and retreats when we remember Allah.",
		LifeApplication: ptrStr("When you have bad thoughts, say A'udhu billahi min ash-shaytan ir-rajeem to push Shaytan away."),
		Topic:           ptrStr("protection"),
		Difficulty:      "medium",
	},
	{
		SurahNumber:     114,
		AyahNumber:      5,
		SurahNameEn:     "An-Nas",
		TextAr:          "الَّذِي يُوَسْوِسُ فِي صُدُورِ النَّاسِ",
		TextEn:          "Who whispers [evil] into the breasts of mankind,",
		Transliteration: ptrStr("Alladhi yuwaswisu fi suduri n-nas"),
		TafsirSimple:    "Shaytan whispers evil suggestions into the hearts of people, trying to make them do wrong things and forget Allah.",
		Topic:           ptrStr("protection"),
		Difficulty:      "medium",
	},
	{
		SurahNumber:     114,
		AyahNumber:      6,
		SurahNameEn:     "An-Nas",
		TextAr:          "مِنَ الْجِنَّةِ وَالنَّاسِ",
		TextEn:          "From among the jinn and mankind.",
		Transliteration: ptrStr("Mina l-jinnati wa n-nas"),
		TafsirSimple:    "Evil whispers can come from both jinn (unseen beings) and from evil humans. We seek Allah's protection from both.",
		Topic:           ptrStr("protection"),
		Difficulty:      "medium",
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

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalf("begin tx: %v", err)
	}

	inserted := 0
	skipped := 0
	for _, v := range verses {
		res, err := tx.ExecContext(ctx,
			`INSERT INTO quran_verses
				(surah_number, ayah_number, surah_name_en, text_ar, text_en,
				 transliteration, tafsir_simple, life_application, topic, difficulty)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			 ON CONFLICT (surah_number, ayah_number) DO NOTHING`,
			v.SurahNumber, v.AyahNumber, v.SurahNameEn, v.TextAr, v.TextEn,
			v.Transliteration, v.TafsirSimple, v.LifeApplication, v.Topic, v.Difficulty,
		)
		if err != nil {
			tx.Rollback()
			log.Fatalf("insert verse %d:%d: %v", v.SurahNumber, v.AyahNumber, err)
		}
		n, _ := res.RowsAffected()
		if n > 0 {
			inserted++
		} else {
			skipped++
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("commit: %v", err)
	}

	fmt.Printf("seeded %d quran verses (%d skipped as duplicates)\n", inserted, skipped)
}

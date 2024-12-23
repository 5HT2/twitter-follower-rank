package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	printer      = message.NewPrinter(language.English)
	w1           = 9
	w2           = 7
	thresholds   = []int{1000000, 100000, 10000, 5000, 1000, 250}
	defaultFile  = "data.json"
	dataRegex    = regexp.MustCompile(`data-[0-9]{4}-[0-9]{2}-[0-9]{2}(-following)?\.json`)
	fileName     = flag.String("f", "", fmt.Sprintf("Data file to read (default \"%s\" or matching \"%s\")", defaultFile, dataRegex.String()))
	modeFollowed = flag.Bool("following", false, "Invert mutuals detection mode to the following tab instead of the followers tab")
	modeRatio    = flag.Bool("ratio", false, "Only display followers with a following:follower ratio of >= -ratioBuf")
	modeRatioBuf = flag.Float64("ratioBuf", 0.9, "Buffer for ranked. e.g. If set to 0.9, it will display if (followers / following >= 0.9)")
	minFollowers = flag.Int("minFollowers", 0, "Filter to >= this many followers")
	maxFollowers = flag.Int("maxFollowers", 0, "Filter to <= this many followers")
	minFollowing = flag.Int("minFollowing", 0, "Filter to >= this many following")
	maxFollowing = flag.Int("maxFollowing", 0, "Filter to <= this many following")
)

type FetchFollowersRange struct {
	Url           string      `json:"url" gorm:"column:url"`
	Content       interface{} `json:"content" gorm:"column:content"`
	contentString string
}

func (f *FetchFollowersRange) UnmarshalJSON(b []byte) error {
	var fJSON FetchFollowersRange
	var cJSON map[string]interface{}

	if err := json.Unmarshal(b, &cJSON); err != nil {
		return err
	}

	contentStr := ""
	if cJSONContent, ok := cJSON["content"]; ok {
		if cJSONContentMap, ok := cJSONContent.(map[string]interface{}); ok {
			if cJSONContentContent, ok := cJSONContentMap["content"]; ok {
				contentStr = fmt.Sprintf("%s", cJSONContentContent)
			} else {
				return fmt.Errorf("fatal: cJSONContentContent") // literally how
			}
		} else {
			contentStr = fmt.Sprintf("%s", cJSONContent)
		}
	} else {
		return fmt.Errorf("fatal: cJSONContent") // probably nil idk
	}

	fJSON = FetchFollowersRange{
		Content:       contentStr,
		contentString: contentStr,
	}
	*f = fJSON
	return nil
}

func (f *FetchFollowersRange) MarshalJSON() ([]byte, error) {
	var u = *f
	return json.MarshalIndent(u, "", "    ")
}

type FetchFollowersContent struct {
	Data struct {
		User struct {
			Result struct {
				Typename string `json:"__typename" gorm:"column:__typename"`
				Timeline struct {
					Timeline struct {
						Instructions []struct {
							Direction string `json:"direction,omitempty"` // New: added between 2024-11-29 → 2024-12-12
							Entries   []struct {
								SortIndex string `json:"sortIndex" gorm:"column:sortIndex"`
								Content   struct {
									EntryType   string `json:"entryType" gorm:"column:entryType"`
									ItemContent struct {
										ItemType    string `json:"itemType" gorm:"column:itemType"`
										Typename    string `json:"__typename" gorm:"column:__typename"`
										UserResults struct {
											Result Follower `json:"result" gorm:"column:result"`
										} `json:"user_results" gorm:"column:user_results"`
										UserDisplayType string `json:"userDisplayType" gorm:"column:userDisplayType"`
									} `json:"itemContent,omitempty" gorm:"column:itemContent"`
									Typename        string `json:"__typename" gorm:"column:__typename"`
									ClientEventInfo struct {
										Component string `json:"component" gorm:"column:component"`
										Element   string `json:"element" gorm:"column:element"`
									} `json:"clientEventInfo,omitempty" gorm:"column:clientEventInfo"`
									Value      string `json:"value,omitempty"`      // New: added between 2024-11-29 → 2024-12-12
									CursorType string `json:"cursorType,omitempty"` // New: added between 2024-11-29 → 2024-12-12
								} `json:"content" gorm:"column:content"`
								EntryID string `json:"entryId" gorm:"column:entryId"`
							} `json:"entries" gorm:"column:entries"`
							Type string `json:"type" gorm:"column:type"`
						} `json:"instructions" gorm:"column:instructions"`
					} `json:"timeline" gorm:"column:timeline"`
				} `json:"timeline" gorm:"column:timeline"`
			} `json:"result" gorm:"column:result"`
		} `json:"user" gorm:"column:user"`
	} `json:"data" gorm:"column:data"`
}

type Follower struct {
	ProfileImageShape string `json:"profile_image_shape" gorm:"column:profile_image_shape"`
	Legacy            struct {
		FriendsCount            int           `json:"friends_count" gorm:"column:friends_count"`
		ProfileImageUrlHttps    string        `json:"profile_image_url_https" gorm:"column:profile_image_url_https"`
		MediaCount              int           `json:"media_count" gorm:"column:media_count"`
		NormalFollowersCount    int           `json:"normal_followers_count" gorm:"column:normal_followers_count"`
		ListedCount             int           `json:"listed_count" gorm:"column:listed_count"`
		DefaultProfileImage     bool          `json:"default_profile_image" gorm:"column:default_profile_image"`
		FavouritesCount         int           `json:"favourites_count" gorm:"column:favourites_count"`
		CreatedAt               string        `json:"created_at" gorm:"column:created_at"`
		Description             string        `json:"description" gorm:"column:description"`
		IsTranslator            bool          `json:"is_translator" gorm:"column:is_translator"`
		WithheldInCountries     []interface{} `json:"withheld_in_countries" gorm:"column:withheld_in_countries"`
		CanMediaTag             bool          `json:"can_media_tag" gorm:"column:can_media_tag"`
		PinnedTweetIDsStr       []string      `json:"pinned_tweet_ids_str" gorm:"column:pinned_tweet_ids_str"`
		FollowedBy              FollowingUser `json:"followed_by" gorm:"column:followed_by"`
		Following               FollowingUser `json:"following" gorm:"column:following"`
		HasCustomTimelines      bool          `json:"has_custom_timelines" gorm:"column:has_custom_timelines"`
		ScreenName              string        `json:"screen_name" gorm:"column:screen_name"`
		WantRetweets            bool          `json:"want_retweets" gorm:"column:want_retweets"`
		TranslatorType          string        `json:"translator_type" gorm:"column:translator_type"`
		CanDm                   bool          `json:"can_dm" gorm:"column:can_dm"`
		PossiblySensitive       bool          `json:"possibly_sensitive" gorm:"column:possibly_sensitive"`
		ProfileInterstitialType string        `json:"profile_interstitial_type" gorm:"column:profile_interstitial_type"`
		Verified                bool          `json:"verified" gorm:"column:verified"`
		Entities                struct {
			Description struct {
				Urls []interface{} `json:"urls" gorm:"column:urls"`
			} `json:"description" gorm:"column:description"`
		} `json:"entities" gorm:"column:entities"`
		StatusesCount      int  `json:"statuses_count" gorm:"column:statuses_count"`
		DefaultProfile     bool `json:"default_profile" gorm:"column:default_profile"`
		FollowersCount     int  `json:"followers_count" gorm:"column:followers_count"`
		FollowersCountStr  string
		FollowingRatioStr  string
		Name               string `json:"name" gorm:"column:name"`
		Location           string `json:"location" gorm:"column:location"`
		FastFollowersCount int    `json:"fast_followers_count" gorm:"column:fast_followers_count"`
	} `json:"legacy" gorm:"column:legacy"`
	HasGraduatedAccess         bool   `json:"has_graduated_access" gorm:"column:has_graduated_access"`
	Typename                   string `json:"__typename" gorm:"column:__typename"`
	IsBlueVerified             bool   `json:"is_blue_verified" gorm:"column:is_blue_verified"`
	ID                         string `json:"id" gorm:"column:id"`
	RestID                     string `json:"rest_id" gorm:"column:rest_id"`
	AffiliatesHighlightedLabel struct {
	} `json:"affiliates_highlighted_label" gorm:"column:affiliates_highlighted_label"`
}

func (f Follower) String() string {
	following := f.Legacy.Following
	if *modeFollowed {
		following = f.Legacy.FollowedBy
	}

	return fmt.Sprintf("%v%s %s %s%s, %s ratio)",
		f.Legacy.FollowersCountStr,
		strings.Repeat(" ", w1-len(f.Legacy.FollowersCountStr)),
		f.Legacy.ScreenName,
		following,
		printer.Sprintf("(%d following", f.Legacy.FriendsCount),
		f.Legacy.FollowingRatioStr,
	)
}

type FollowingUser bool

func (f FollowingUser) String() string {
	if f {
		return "(mutuals) "
	} else {
		return ""
	}
}

func main() {
	flag.Parse()

	dataFiles := findFiles()
	d, dfn, err := unsafeReadFiles(dataFiles)
	if err != nil {
		log.Panicf("%v\n", err)
	} else {
		log.Printf("Data from: \"%s\"\n", dfn)
	}

	followersRange := make([]FetchFollowersRange, 0)
	if err := json.Unmarshal(d, &followersRange); err != nil {
		log.Panicf("%v\n", err)
	}

	followers := make([]Follower, 0)
	for _, r := range followersRange {
		//jRaw, err := strconv.Unquote(r.Content.Content)
		//if err != nil {
		//	log.Panicf("===\n%s\n===\n%s\n===\n%v\n", r.Content.Content, jRaw, err)
		//}

		// Allow skipping rate-limited error responses that aren't JSON
		if len(r.contentString) == 0 || r.contentString[0] != '{' {
			continue
		}

		followersContent := FetchFollowersContent{}
		if err := json.Unmarshal([]byte(r.contentString), &followersContent); err != nil {
			log.Panicf("%v\n", err)
		}

		for _, d1 := range followersContent.Data.User.Result.Timeline.Timeline.Instructions {
			if d1.Type == "TimelineAddEntries" {
				for _, entry := range d1.Entries {
					user := entry.Content.ItemContent.UserResults.Result.Legacy

					if len(user.ScreenName) == 0 { // no account info to display
						continue
					}

					numFollow := user.FollowersCount
					numFriend := user.FriendsCount

					// Filter to specified ranges of follower counts
					if numFollow < *minFollowers {
						continue
					}
					if numFollow > *maxFollowers && *maxFollowers > 0 {
						continue
					}
					if numFriend < *minFollowing {
						continue
					}
					if numFriend > *maxFollowing && *maxFollowing > 0 {
						continue
					}

					// Filter to follower:following ratio
					userRatio := float64(numFollow) / float64(numFriend) // (followers / following)
					passRatio := userRatio > *modeRatioBuf

					// If a negative ratio is specified, we instead want to
					// filter to following:follower ratio
					if *modeRatioBuf < 0.0 {
						userRatio = -1.0 * (float64(numFriend) / float64(numFollow)) // -1 * (following / followers)
						passRatio = userRatio < *modeRatioBuf
					}

					if *modeRatio && !passRatio {
						continue
					}

					user.FollowersCountStr = printer.Sprintf("%d", user.FollowersCount)
					user.FollowingRatioStr = fmt.Sprintf("%.3f", userRatio)
					entry.Content.ItemContent.UserResults.Result.Legacy = user

					followers = append(followers, entry.Content.ItemContent.UserResults.Result)
				}
			}
		}
	}

	sort.SliceStable(followers, func(i, j int) bool {
		return followers[i].Legacy.FollowersCount < followers[j].Legacy.FollowersCount
	})

	w1 = len(followers[len(followers)-1].Legacy.FollowersCountStr)
	numFollowers := make([]int, len(thresholds))

	for _, user := range followers {
		for k, v := range thresholds {
			if user.Legacy.FollowersCount >= v {
				numFollowers[k]++
				break // Break after incrementing the first applicable threshold
			}
		}

		log.Printf("%s\n", user)
	}

	w2 = len(printer.Sprintf("%d", thresholds[0])) + 1
	log.Printf("Followers with more than:")

	for i := len(numFollowers) - 1; i >= 0; i-- {
		if numFollowers[i] == 0 {
			continue
		}

		threshold := printer.Sprintf("%d", thresholds[i])
		log.Printf(
			"  %s:%s%d",
			threshold,
			strings.Repeat(" ", w2-len(threshold)),
			numFollowers[i],
		)
	}
}

func findFiles() []string {
	files := make([]string, 0)
	f := *fileName

	// First try the user-specified file
	if f != "" {
		files = append(files, f)
	}

	// Now we want to find files matching dataRegex, in order of most recent
	if de, err := os.ReadDir("."); err != nil {
		return files
	} else {
		filesUnsorted := make([]string, 0)

		for _, e := range de {
			// Don't continue on directories or files not matching the regex
			if e.IsDir() || !dataRegex.MatchString(e.Name()) {
				continue
			}

			em := dataRegex.FindStringSubmatch(e.Name())
			if len(em) != 2 { // Shouldn't ever happen, safety check
				continue
			}

			// If we have modeFollowed enabled we want to require -following in the filename
			if *modeFollowed == (em[1] == "-following") {
				filesUnsorted = append(filesUnsorted, em[0])
			}
		}

		slices.Sort(filesUnsorted)
		slices.Reverse(filesUnsorted)
		files = append(files, filesUnsorted...)
	}

	// Lastly, try the default file, if we haven't already added it
	if f != defaultFile {
		files = append(files, defaultFile)
	}

	return files
}

// unsafeReadFiles will read a slice of filenames until it successfully opens one, and returns its contents.
// It will return the very first error it encountered if it does not find a file.
func unsafeReadFiles(files []string) ([]byte, string, error) {
	var err error
	var efn string

	for n, s := range files {
		if d, errTmp := os.ReadFile(s); errTmp == nil {
			return d, s, nil
		} else {
			if n == 0 {
				err = errTmp
				efn = s
			}

			log.Printf("unsafeReadFiles: %s: %s\n", efn, err)
		}
	}

	return nil, efn, err
}

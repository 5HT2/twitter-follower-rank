package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"log"
	"os"
	"sort"
	"strings"
)

var (
	printer      = message.NewPrinter(language.English)
	w1           = 9
	w2           = 7
	thresholds   = []int{1000000, 100000, 10000, 5000, 1000, 250}
	fileName     = flag.String("f", "data.json", "Data file to read")
	modeFollowed = flag.Bool("following", false, "Invert mutuals detection mode to the following tab instead of the followers tab")
)

type FetchFollowersRange struct {
	Url     string `json:"url" gorm:"column:url"`
	Content struct {
		Error   interface{} `json:"error" gorm:"column:error"`
		Content string      `json:"content" gorm:"column:content"`
		Encoded bool        `json:"encoded" gorm:"column:encoded"`
	} `json:"content" gorm:"column:content"`
}

type FetchFollowersContent struct {
	Data struct {
		User struct {
			Result struct {
				Typename string `json:"__typename" gorm:"column:__typename"`
				Timeline struct {
					Timeline struct {
						Instructions []struct {
							Entries []struct {
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
									} `json:"itemContent" gorm:"column:itemContent"`
									Typename        string `json:"__typename" gorm:"column:__typename"`
									ClientEventInfo struct {
										Component string `json:"component" gorm:"column:component"`
										Element   string `json:"element" gorm:"column:element"`
									} `json:"clientEventInfo" gorm:"column:clientEventInfo"`
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

	return fmt.Sprintf("%v%s %s %s%s",
		f.Legacy.FollowersCountStr,
		strings.Repeat(" ", w1-len(f.Legacy.FollowersCountStr)),
		f.Legacy.ScreenName,
		following,
		printer.Sprintf("(%d following)", f.Legacy.FriendsCount),
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

	d, err := os.ReadFile(*fileName)
	if err != nil {
		log.Panicf("%v\n", err)
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
		if len(r.Content.Content) == 0 || r.Content.Content[0] != '{' {
			continue
		}

		followersContent := FetchFollowersContent{}
		if err := json.Unmarshal([]byte(r.Content.Content), &followersContent); err != nil {
			log.Panicf("%v\n", err)
		}

		for _, d1 := range followersContent.Data.User.Result.Timeline.Timeline.Instructions {
			if d1.Type == "TimelineAddEntries" {
				for _, entry := range d1.Entries {
					entry.Content.ItemContent.UserResults.Result.Legacy.FollowersCountStr = printer.Sprintf("%d", entry.Content.ItemContent.UserResults.Result.Legacy.FollowersCount)
					followers = append(followers, entry.Content.ItemContent.UserResults.Result)
				}
			} else {
				log.Printf("Skipping: %v\n", d1)
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

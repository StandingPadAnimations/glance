package feed

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type feedbinGetEntryResponseJson struct {
	Id int `json:"id"`
	FeedId int `json:"feed_id"`
	Title *string `json:"title"`
	Url string `json:"url"`
	ExtractedContentUrl string `json:"extracted_content_url"`
	Author *string `json:"author"`
	Content *string `json:"content"`
	Summary string `json:"summary"`
	Published string `json:"published"`
	CreatedAt string `json:"created_at"`
}

type feedbinGetEntryResponseJsonList []feedbinGetEntryResponseJson

type feedbinFeedResponseJson struct {
	Id int `json:"id"`
	Title string `json:"title"`
	FeedUrl string `json:"feed_url"`
	SiteUrl string `json:"site_url"`
}

func getFeedBinUnreadPostIds(username string, password string) ([]int, error) {
	request, _ := http.NewRequest("GET", "https://api.feedbin.com/v2/unread_entries.json", nil)
	request.SetBasicAuth(username, password)
	response, err := decodeJsonFromRequest[[]int](defaultClient, request)
	
	if err != nil {
		return nil, fmt.Errorf("%w: could not fetch list of post IDs", ErrNoContent)
	}

	return response, nil
}


func convertTimeToTimestamp(date string) int64 {
	format := "2006-01-02T15:04:05.000000Z"

	t, err := time.Parse(format, date)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(t.Unix())
		return int64(t.Unix())
	}
	return 0
}

func getFeedBinUnreadPostsFromIds(postIds []int, username string, password string) (FeedBinPosts, error) {
	ids := strings.Trim(strings.Replace(fmt.Sprint(postIds), " ", ",", -1), "[]")
	request, _ := http.NewRequest("GET", fmt.Sprintf("https://api.feedbin.com/v2/entries.json?ids=%s", ids), nil)
	request.SetBasicAuth(username, password)

	results, err := decodeJsonFromRequest[feedbinGetEntryResponseJsonList](defaultClient, request)

	if err != nil {
		return nil, err
	}

	posts := make(FeedBinPosts, 0, len(postIds))

	feed_providers := make(map[int]string)
	for i := range results {
		title := ""

		if results[i].Title != nil {
			title = *results[i].Title
		}

		// Get the feed provider. This performs 
		// a request if and only if we haven't gotten 
		// the provider yet
		provider, exists := feed_providers[results[i].FeedId]
		feed := ""
		if exists {
			feed = provider
		} else {
			feed_request, _ := http.NewRequest("GET", fmt.Sprintf("https://api.feedbin.com/v2/feeds/%d.json", results[i].FeedId), nil)
			feed_request.SetBasicAuth(username, password)

			feed_result, feed_err := decodeJsonFromRequest[feedbinFeedResponseJson](defaultClient, feed_request)
			
			if feed_err == nil {
				feed = feed_result.Title
				feed_providers[results[i].FeedId] = feed_result.Title
			}
		}

		posts = append(posts, FeedBinPost{
			Title:           title,
			Feed:		 feed,
			FeedId:		 results[i].FeedId,
			TargetUrl:       fmt.Sprintf("https://feedbin.com/entries/%d", results[i].Id),
			TimePosted:      time.Unix(convertTimeToTimestamp(results[i].Published), 0),
		})
	}

	if len(posts) == 0 {
		return nil, ErrNoContent
	}

	if len(posts) != len(postIds) {
		return posts, fmt.Errorf("%w could not fetch some Feedbin posts", ErrPartialContent)
	}

	return posts, nil
}

func FetchFeedbinUnreadPosts(limit int, username string, password string) (FeedBinPosts, error) {
	postIds, err := getFeedBinUnreadPostIds(username, password)

	if err != nil {
		return nil, err
	}

	if len(postIds) > limit {
		postIds = postIds[:limit]
	}

	return getFeedBinUnreadPostsFromIds(postIds, username, password)
}

func FetchFeedbinUnreadFeeds(limit int, username string, password string) (FeedBinFeeds, error) {
	posts, err := FetchFeedbinUnreadPosts(limit, username, password)

	if err != nil {
		return nil, err
	}

	feeds_map := make(map[int]FeedBinFeed)

	for _, p := range posts {
		post := p
		val, exists := feeds_map[post.FeedId]
		if exists {
			newPosts := make(FeedBinPosts, len(val.Posts))
			copy(newPosts, val.Posts)
			newPosts = append(newPosts, post)
			val.Posts = newPosts
			feeds_map[post.FeedId] = val
		} else {
			feeds_map[post.FeedId] = FeedBinFeed {
				Posts: FeedBinPosts{post},
				FeedId: post.FeedId,
				FeedName: post.Feed,
			}
		}
	}
	
	feeds_arr := make(FeedBinFeeds, 0, len(feeds_map))
	for _, value := range feeds_map {
		feeds_arr = append(feeds_arr, value)
	}
	return feeds_arr, nil
}

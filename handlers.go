package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/l4b4r4b4b4/arxiv-rss-assistant/internal/database"
)

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	responseWithJSON(w, 200, map[string]string{"status": "ok"})
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	responseWithError(w, 400, "Something went wrong")
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name string `name`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		responseWithError(w, 400, fmt.Sprintf("Can't decode request body: %v", err))
		return
	}
	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
	})
	if err != nil {
		responseWithError(w, 400, fmt.Sprintf("Could not create user: %v", err))
	}

	responseWithJSON(w, 201, databaseUserToUser(user))
}

func (apiConfig *apiConfig) getUserHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	responseWithJSON(w, http.StatusOK, databaseUserToUser(user))

}

func (cfg *apiConfig) createFeedHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	feed, err := cfg.DB.CreateFeed(r.Context(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		Name:      params.Name,
		Url:       params.URL,
	})
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't create feed")
		return
	}

	responseWithJSON(w, http.StatusCreated, databaseFeedToFeed(feed))
}

func (cfg *apiConfig) getFeedsHandler(w http.ResponseWriter, r *http.Request) {
	feeds, err := cfg.DB.GetFeeds(r.Context())
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't get feeds: %v", err))
		return
	}

	responseWithJSON(w, http.StatusCreated, databaseFeedsToFeeds(feeds))
}

func (cfg *apiConfig) createFeedFollowHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		FeedId uuid.UUID `json:"feed_id"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	feedFollow, err := cfg.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    params.FeedId,
	})
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't create feed follow: %v", err))
		return
	}

	responseWithJSON(w, http.StatusCreated, databaseFeedFollowToFeedFollow(feedFollow))
}

func (cfg *apiConfig) getFeedFollowsHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	feed_follows, err := cfg.DB.GetFeedFollows(r.Context(), user.ID)
	if err != nil {
		responseWithError(w, 400, fmt.Sprintf("Couldn't get feed follows: %v", err))
		return
	}

	responseWithJSON(w, http.StatusOK, databaseFeedFollowsToFeedFollows(feed_follows))
}

func (cfg *apiConfig) deleteFeedFollowsHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	feedFollowIDStr := chi.URLParam(r, "feedFollowID")
	feedFollowID, err := uuid.Parse(feedFollowIDStr)
	if err != nil {
		responseWithError(w, 400, fmt.Sprintf("Couldn't delete feed follow ID: %v", err))
		return
	}
	cfg.DB.DeleteFeedFollow(r.Context(), database.DeleteFeedFollowParams{
		ID:     feedFollowID,
		UserID: user.ID,
	})
	responseWithJSON(w, http.StatusOK, map[string]string{"delete": "successful"})
}

func (cfg *apiConfig) getPostsForUserHandler(w http.ResponseWriter, r *http.Request, user database.User) {
posts, err :=	cfg.DB.GetPostsForUser(r.Context(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  10,
	})
	if err != nil {
		responseWithError(w, 400, fmt.Sprintf("Couldn't get posts for user: %v", err))
		return
	}
	responseWithJSON(w, http.StatusOK, databasePostsToPosts(posts))
}

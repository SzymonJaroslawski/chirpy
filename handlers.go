package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/SzymonJaroslawski/chirpy/internal/auth"
	"github.com/SzymonJaroslawski/chirpy/internal/database"
	"github.com/google/uuid"
)

var PROFANE = []string{"kerfuffle", "sharbert", "fornax"}

type Chirp struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	Id        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
}

type User struct {
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	Id           uuid.UUID `json:"id"`
}

func handlePutUsers(w http.ResponseWriter, r *http.Request, cfg *apiConfig) {
	type Parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	tokenHeaderValue, err := auth.GetBearerToken(r.Header.Clone())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	tokenId, err := auth.ValidateJWT(tokenHeaderValue, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	newHashedPasswd, err := auth.HashedPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	user, err := cfg.db.UpdateUserEmailAndPassword(context.Background(), database.UpdateUserEmailAndPasswordParams{
		Email:          params.Email,
		HashedPassowrd: newHashedPasswd,
		ID:             tokenId,
	})

	type Response struct {
		Email string `json:"email"`
	}

	res := Response{
		Email: user.Email,
	}

	err = respondWithJSON(w, http.StatusOK, res)
	if err != nil {
		log.Printf("Error sending response: %s", err.Error())
	}
}

func handleRevoke(w http.ResponseWriter, r *http.Request, cfg *apiConfig) {
	token, err := auth.GetBearerToken(r.Header.Clone())
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	tokenDB, err := cfg.db.GetRefreshToken(context.Background(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	err = cfg.db.RevokeRefreshToken(context.Background(), tokenDB.Token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleRefresh(w http.ResponseWriter, r *http.Request, cfg *apiConfig) {
	token, err := auth.GetBearerToken(r.Header.Clone())
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	tokenDB, err := cfg.db.GetRefreshToken(context.Background(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if time.Now().UTC().After(tokenDB.ExpiresAt) || tokenDB.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "token expired")
		return
	}

	newToken, err := auth.MakeJWT(tokenDB.UserID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	type Resposne struct {
		Token string `json:"token"`
	}

	res := Resposne{
		Token: newToken,
	}

	err = respondWithJSON(w, http.StatusOK, res)
	if err != nil {
		log.Printf("Error sending response: %s", err.Error())
	}
}

func handleMetrics(w http.ResponseWriter, r *http.Request, cfg *apiConfig) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	html := `
    <html>
      <body>
        <h1>Welcome, Chirpy Admin</h1>
        <p>Chirpy has been visited %d times!</p>
      </body>
    </html>
    `

	fmt.Fprintf(w, html, cfg.fileserverHits.Load())
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleReset(w http.ResponseWriter, r *http.Request, cfg *apiConfig) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	cfg.fileserverHits.Store(0)
	err := cfg.db.ResetUsers(context.Background())
	if err != nil {
		log.Printf("Error reseting user database: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Error       string `json:"error,omitempty"`
		CleanedBody string `json:"cleaned_body,omitempty"`
		Valid       bool   `json:"valid,omitempty"`
	}

	type Parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	code := http.StatusOK

	res := Response{
		Valid:       true,
		CleanedBody: validateProfaneLogic(params.Body, PROFANE),
	}

	if len(params.Body) > 140 {
		res = Response{
			Error: "Chirp is too long",
		}

		code = http.StatusBadRequest
	}

	dat, err := json.Marshal(res)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func handleGetChirp(w http.ResponseWriter, r *http.Request, cfg *apiConfig) {
	chirp_id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("Error parsing uuid: %s", err)
		err = respondWithError(w, http.StatusBadRequest, "Invalid id")
		if err != nil {
			log.Printf("Error sending response: %s", err)
			return
		}
		return
	}
	chirp, err := cfg.db.GetChirpWithId(context.Background(), chirp_id)
	if err != nil {
		log.Printf("Error geting chirp with id: %s, %s", chirp_id, err)
		err = respondWithError(w, http.StatusNotFound, "Chrip not found")
		if err != nil {
			log.Printf("Error sending error response: %s", err)
			return
		}
		return
	}

	res_chirp := Chirp{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	err = respondWithJSON(w, http.StatusOK, res_chirp)
	if err != nil {
		log.Printf("Error sending resposne: %s", err)
		return
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request, cfg *apiConfig) {
	type Parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding JSON: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error decoding json")
		return
	}

	user, err := cfg.db.GetUserWithEmail(context.Background(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassowrd)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Wrong password")
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Duration(time.Hour))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = cfg.db.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 1440),
	})

	res := User{
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Id:           user.ID,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	}

	respondWithJSON(w, http.StatusOK, res)
}

func handleGetAllChirps(w http.ResponseWriter, r *http.Request, cfg *apiConfig) {
	chirps, err := cfg.db.GetAllChirps(context.Background())
	if err != nil {
		log.Printf("Error retriving chirps: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "retriving chirps")
		if err != nil {
			log.Printf("Error sending error response: %s", err)
			return
		}
		return
	}

	var chirpsRes []Chirp

	for _, chirp := range chirps {
		chirpsRes = append(chirpsRes, Chirp{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	err = respondWithJSON(w, http.StatusOK, chirpsRes)
	if err != nil {
		log.Printf("Error sending resposne: %s", err)
		return
	}
}

func handleCreateChirp(w http.ResponseWriter, r *http.Request, cfg *apiConfig) {
	type Response struct {
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
		Id        uuid.UUID `json:"id"`
	}

	type Parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding params: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Internal Server Error: "+strconv.Itoa(http.StatusInternalServerError))
		if err != nil {
			log.Printf("Error sending error response: %s", err)
			return
		}
		return
	}

	tokenHeader, err := auth.GetBearerToken(r.Header.Clone())
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
	}

	tokenID, err := auth.ValidateJWT(tokenHeader, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	if len(params.Body) > 140 {
		err = respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		if err != nil {
			log.Printf("Error sending error response: %s", err)
		}
		return
	}

	chirp, err := cfg.db.CreateChirp(context.Background(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: tokenID,
	})

	res := Response{
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
		Id:        chirp.ID,
	}

	err = respondWithJSON(w, http.StatusCreated, res)
	if err != nil {
		log.Printf("Error sending response: %s", err)
	}
}

func handleCreateUser(w http.ResponseWriter, r *http.Request, cfg *apiConfig) {
	type Response struct {
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Id        uuid.UUID `json:"id"`
	}

	type Parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding params: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if params.Password == "" || len(params.Password) < 5 {
		err = respondWithError(w, http.StatusBadRequest, "Wrong password lenght")
		if err != nil {
			log.Printf("Error sending error response: %s", err)
			return
		}
		return
	}

	hashed, err := auth.HashedPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "While hashing password")
		if err != nil {
			log.Printf("Error sending error response: %s", err)
		}
		return
	}

	user, err := cfg.db.CreateUser(context.Background(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassowrd: hashed,
	})
	if err != nil {
		log.Printf("Error creating user: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res := Response{
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Id:        user.ID,
	}

	err = respondWithJSON(w, http.StatusCreated, res)
	if err != nil {
		log.Printf("Error sending json: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var (
	ctx = context.Background()
	rdb *redis.Client
)

// =======================
// REQUEST STRUCT
// =======================
type ReviewRequest struct {
	CourseID string `json:"course_id"`
	UserID   string `json:"user_id"`
	Rating   int    `json:"rating"`
	Comment  string `json:"comment"`
}

func main() {
	// =======================
	// REDIS CONNECTION
	// =======================
	rdb = redis.NewClient(&redis.Options{
		Addr: "review-redis:6379", // NAMA SERVICE REDIS DI DOCKER
	})

	r := gin.Default()

	// =======================
	// CREATE REVIEW
	// =======================
	r.POST("/reviews", func(c *gin.Context) {
		var body ReviewRequest
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := rdb.Incr(ctx, "review:id").Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		review := map[string]interface{}{
			"id":         id,
			"course_id":  body.CourseID,
			"user_id":    body.UserID,
			"rating":     body.Rating,
			"comment":    body.Comment,
			"created_at": time.Now().Format(time.RFC3339),
		}

		reviewKey := "review:" + strconv.FormatInt(id, 10)

		// simpan review
		rdb.HSet(ctx, reviewKey, review)

		// index by course
		rdb.SAdd(ctx, "course:"+body.CourseID+":reviews", id)

		// index by user (INI YANG PENTING)
		rdb.SAdd(ctx, "user:"+body.UserID+":reviews", id)

		c.JSON(http.StatusCreated, review)
	})

	// =======================
	// GET REVIEWS BY COURSE
	// =======================
	r.GET("/reviews/course/:course_id", func(c *gin.Context) {
		courseID := c.Param("course_id")

		ids, _ := rdb.SMembers(ctx, "course:"+courseID+":reviews").Result()

		var reviews []map[string]string
		for _, id := range ids {
			data, _ := rdb.HGetAll(ctx, "review:"+id).Result()
			if len(data) > 0 {
				reviews = append(reviews, data)
			}
		}

		c.JSON(http.StatusOK, reviews)
	})

	// =======================
	// GET REVIEWS BY USER
	// =======================
	r.GET("/reviews/user/:user_id", func(c *gin.Context) {
		userID := c.Param("user_id")

		ids, _ := rdb.SMembers(ctx, "user:"+userID+":reviews").Result()

		var reviews []map[string]string
		for _, id := range ids {
			data, _ := rdb.HGetAll(ctx, "review:"+id).Result()
			if len(data) > 0 {
				reviews = append(reviews, data)
			}
		}

		c.JSON(http.StatusOK, reviews)
	})

	// =======================
	// UPDATE REVIEW
	// =======================
	r.PUT("/reviews/:id", func(c *gin.Context) {
		id := c.Param("id")

		var body ReviewRequest
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		rdb.HSet(ctx, "review:"+id, map[string]interface{}{
			"rating":  body.Rating,
			"comment": body.Comment,
		})

		c.JSON(http.StatusOK, gin.H{"message": "updated"})
	})

	// =======================
	// DELETE REVIEW
	// =======================
	r.DELETE("/reviews/:id", func(c *gin.Context) {
		id := c.Param("id")

		courseID, _ := rdb.HGet(ctx, "review:"+id, "course_id").Result()
		userID, _ := rdb.HGet(ctx, "review:"+id, "user_id").Result()

		rdb.Del(ctx, "review:"+id)
		rdb.SRem(ctx, "course:"+courseID+":reviews", id)
		rdb.SRem(ctx, "user:"+userID+":reviews", id)

		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})
// =======================
	// ADMIN: GET ALL REVIEWS
	// =======================
	r.GET("/admin/reviews", func(c *gin.Context) {
		
		maxIDStr, err := rdb.Get(ctx, "review:id").Result()
		if err == redis.Nil {
			c.JSON(http.StatusOK, []interface{}{}) 
			return
		}

		maxID, _ := strconv.Atoi(maxIDStr)
		var allReviews []map[string]string

		// 2. Loop dari 1 sampai Max ID
		for i := 1; i <= maxID; i++ {
			key := "review:" + strconv.Itoa(i)
			
			// 3. Cek apakah key review ini ada (mungkin sudah dihapus)
			exists, _ := rdb.Exists(ctx, key).Result()
			
			if exists > 0 {
				// 4. Ambil datanya
				data, _ := rdb.HGetAll(ctx, key).Result()
				allReviews = append(allReviews, data)
			}
		}

		// Balikan semua data
		c.JSON(http.StatusOK, allReviews)
	})
	// =======================
	// ADMIN: GET ALL REVIEWS
	// =======================
	r.GET("/admin/reviews", func(c *gin.Context) {
		maxIDStr, err := rdb.Get(ctx, "review:id").Result()
		if err == redis.Nil {
			c.JSON(http.StatusOK, []interface{}{})
			return
		}

		maxID, _ := strconv.Atoi(maxIDStr)
		var reviews []map[string]string

		for i := 1; i <= maxID; i++ {
			key := "review:" + strconv.Itoa(i)
			exists, _ := rdb.Exists(ctx, key).Result()
			if exists > 0 {
				data, _ := rdb.HGetAll(ctx, key).Result()
				reviews = append(reviews, data)
			}
		}

		c.JSON(http.StatusOK, reviews)
	})

	// =======================
	// RUN SERVER
	// =======================
	r.Run(":8080")
}

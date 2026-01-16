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
		Addr: "review-redis:6379",
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

		id, _ := rdb.Incr(ctx, "review:id").Result()

		review := map[string]interface{}{
			"id":         id,
			"course_id":  body.CourseID,
			"user_id":    body.UserID,
			"rating":     body.Rating,
			"comment":    body.Comment,
			"created_at": time.Now().Format(time.RFC3339),
		}

		key := "review:" + strconv.FormatInt(id, 10)
		rdb.HSet(ctx, key, review)
		rdb.SAdd(ctx, "course:"+body.CourseID+":reviews", id)

		c.JSON(http.StatusCreated, review)
	})

	// =======================
	// READ BY COURSE
	// =======================
	r.GET("/reviews/:course_id", func(c *gin.Context) {
		courseID := c.Param("course_id")
		ids, _ := rdb.SMembers(ctx, "course:"+courseID+":reviews").Result()

		var reviews []map[string]string
		for _, id := range ids {
			data, _ := rdb.HGetAll(ctx, "review:"+id).Result()
			reviews = append(reviews, data)
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

		rdb.Del(ctx, "review:"+id)
		rdb.SRem(ctx, "course:"+courseID+":reviews", id)

		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// =======================
	// RUN SERVER
	// =======================
	r.Run(":8080")
}

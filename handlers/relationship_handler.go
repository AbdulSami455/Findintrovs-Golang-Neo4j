package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func CreateSimpleRelationshipHandler(c *gin.Context, driver neo4j.Driver) {
	var input struct {
		Person1      string `json:"person1"`
		Person2      string `json:"person2"`
		Relationship string `json:"relationship"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	query := `
		MATCH (a:User {username: $person1}), (b:User {username: $person2})
		CREATE (a)-[:` + input.Relationship + `]->(b)
	`

	params := map[string]interface{}{
		"person1": input.Person1,
		"person2": input.Person2,
	}

	_, err := session.Run(query, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create relationship: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Relationship created successfully"})
}

func MatchPreferencesAndCreateRelationship(c *gin.Context, driver neo4j.Driver) {
	var input struct {
		Person1 string `json:"person1"`
		Person2 string `json:"person2"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	query := `
		MATCH (a:User {username: $person1}), (b:User {username: $person2})
		WITH a, b,
			[
				size([x IN a.movies_likes WHERE x IN b.movies_likes]),
				size([x IN a.movies_dislikes WHERE x IN b.movies_dislikes]),
				size([x IN a.games_likes WHERE x IN b.games_likes]),
				size([x IN a.games_dislikes WHERE x IN b.games_dislikes]),
				size([x IN a.books_likes WHERE x IN b.books_likes]),
				size([x IN a.books_dislikes WHERE x IN b.books_dislikes]),
				size([x IN a.music_likes WHERE x IN b.music_likes]),
				size([x IN a.music_dislikes WHERE x IN b.music_dislikes]),
				size([x IN a.art_hobbies WHERE x IN b.art_hobbies]),
				size([x IN a.outdoors_likes WHERE x IN b.outdoors_likes]),
				size([x IN a.outdoors_dislikes WHERE x IN b.outdoors_dislikes]),
				size([x IN a.fitness_hobbies WHERE x IN b.fitness_hobbies]),
				size([x IN a.social_hobbies WHERE x IN b.social_hobbies])
			] AS match_counts
		WITH a, b, reduce(total = 0, count IN match_counts | total + count) AS total_matches
		CREATE (a)-[r:SIMILAR_PREFERENCES {matches: total_matches}]->(b)
		RETURN r.matches AS match_count
	`

	params := map[string]interface{}{
		"person1": input.Person1,
		"person2": input.Person2,
	}

	result, err := session.Run(query, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process preferences: " + err.Error()})
		return
	}

	if result.Next() {
		matchCount, _ := result.Record().Get("match_count")
		c.JSON(http.StatusOK, gin.H{
			"message":     "Relationship created based on preferences",
			"match_count": matchCount,
		})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Users not found or no matches"})
	}
}

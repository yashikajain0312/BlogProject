package main

import (
	"github.com/gin-gonic/gin"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "fmt"
    "net/http"
    "context"
    "blogging/ent"
    "blogging/ent/post"
    "strconv"
    "log"
)

var db *sql.DB
var client *ent.Client

func main() {

    dbConnection()

     // Initialize the Ent client
     var err error
     client, err = ent.Open("mysql", "root:@tcp(localhost:3306)/blog?parseTime=true")
     if err != nil {
         log.Fatalf("Error opening Ent client: %v", err)
     }
     defer client.Close()
    
	r := gin.Default()

    r.GET("/blog/posts/get", BlogGET)
    r.GET("/blog/post/:id/get", SpecificBlogGET)
    r.POST("/blog/post/create", BlogPOST)    
    r.PUT("/blog/post/:id/update", BlogPUT)
    r.DELETE("/blog/post/:id/delete", BlogDELETE)

	if err := r.Run(":9000"); err != nil {
        log.Fatalf("failed to start the server: %v", err)
	}
}

func dbConnection() {
    var err error
        db, err = sql.Open("mysql", "root:@tcp(localhost:3306)/blog?parseTime=true")
    if err != nil {
        log.Fatalf("Error opening database connection: %v", err)
    }
    defer db.Close()
    
    err = db.Ping()
    if err != nil {
        log.Fatalf("Error pinging database: %v", err)
    }
    
    fmt.Println("Connected to the database!")
    
    // Creating table
    _, err = db.Exec(`
            CREATE TABLE IF NOT EXISTS posts (
                id INT AUTO_INCREMENT PRIMARY KEY,
                title VARCHAR(255) NOT NULL,
                content TEXT NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
            )
        `)
    if err != nil {
        log.Fatalf("Error creating table in database: %v", err)
    }
    
    fmt.Println("Table 'posts' created successfully!")
    }

// task 1: Retrieve a list of all blog posts.
func BlogGET(c *gin.Context) {

    posts, err := client.Post.Query().All(context.Background())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve posts"})
        return
    }

    var response []gin.H
    for _, p := range posts {
        response = append(response, gin.H{
            "id":      p.ID,
            "title":   p.Title,
            "content": p.Content,
        })
    }

    c.JSON(http.StatusOK, response)
}

// task 2: Retrieve a specific blog post by ID.
func SpecificBlogGET(c *gin.Context) {

    var getRequestURI struct {
        ID int `uri:"id" binding:"required"`
    }

    if err := c.ShouldBindUri(&getRequestURI); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    post, err := client.Post.Query().
        Where(post.IDEQ(getRequestURI.ID)).
        Only(context.Background())
    if err != nil {
        if ent.IsNotFound(err) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Post not found with ID: " + strconv.Itoa(getRequestURI.ID)})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve post from database"})
        }
        return
    }

    c.JSON(http.StatusOK, post)
}

// task 3: Create a new blog post.
func BlogPOST(c *gin.Context) {
    var postRequest struct {
        Title   string `json:"title" binding:"required"`
        Content string `json:"content" binding:"required"`
    }

    if err := c.ShouldBindJSON(&postRequest); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    newPost, err := client.Post.Create().
        SetTitle(postRequest.Title).
        SetContent(postRequest.Content).
        Save(context.Background())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "Post created successfully!", "post": newPost})
}

// task 4: Update an existing blog post.
func BlogPUT(c *gin.Context) {
    var updateRequest struct {
        Title   string `json:"title"`
        Content string `json:"content"`
    }

    var updateRequestURI struct {
        ID int `uri:"id" binding:"required"`
    }

    if err := c.ShouldBindJSON(&updateRequest); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := c.ShouldBindUri(&updateRequestURI); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    post, err := client.Post.Query().
        Where(post.IDEQ(updateRequestURI.ID)).
        Only(context.Background())
    if err != nil {
        if ent.IsNotFound(err) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Post not found with ID: " + strconv.Itoa(updateRequestURI.ID)})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve post from database"})
        }
        return
    }

    update := post.Update()

    if updateRequest.Title != "" {
        update = update.SetTitle(updateRequest.Title)
    }

    if updateRequest.Content != "" {
        update = update.SetContent(updateRequest.Content)
    }

    updatedPost, err := update.Save(context.Background())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
        return
    }
        
        
    c.JSON(http.StatusOK, updatedPost)
}

// task 5: Delete a blog post.
func BlogDELETE(c *gin.Context) {

    var deleteRequestURI struct {
        ID int `uri:"id" binding:"required"`
    }

    if err := c.ShouldBindUri(&deleteRequestURI); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := client.Post.DeleteOneID(deleteRequestURI.ID).Exec(context.Background()); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}


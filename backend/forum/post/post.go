package post

func (h *PostHandler) ServePostsPage(c *gin.Context) {
    userID, ok := getUserID(c)
    if !ok {
        c.Redirect(http.StatusFound, "/login")
        return
    }

    posts, err := h.DB.GetPosts(postdb.PostFilter{}, userID)
    if err != nil {
        c.HTML(http.StatusInternalServerError, "500.html", gin.H{
            "error": "Impossible de charger les posts",
        })
        return
    }

    c.HTML(http.StatusOK, "posts.html", gin.H{
        "Posts": posts,
    })
}
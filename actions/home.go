package actions

import (
	"fmt"
	"io"
	"time"

	"github.com/bscott/golangflow/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/pop/v5"
	"github.com/gorilla/feeds"
	"github.com/pkg/errors"
	stripmd "github.com/writeas/go-strip-markdown"
)

// HomeHandler is a default handler to serve up
// a home page.
func HomeHandler(c buffalo.Context) error {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>GolangFlow - Maintenance</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://use.fontawesome.com/releases/v5.0.13/css/all.css">
</head>
<body class="bg-light">
    <div class="container">
        <div class="row justify-content-center mt-5">
            <div class="col-md-8 text-center">
                <h1 class="display-4 mb-4">🔧 GolangFlow</h1>
                <div class="card shadow">
                    <div class="card-body p-5">
                        <h2 class="card-title">Under Maintenance</h2>
                        <p class="card-text lead">We're currently upgrading our systems to serve you better.</p>
                        <p class="card-text">The site will be back online shortly. Thank you for your patience!</p>
                        <hr>
                        <div class="d-flex justify-content-center gap-3">
                            <a href="/rss" class="btn btn-outline-primary">
                                <i class="fas fa-rss"></i> RSS Feed
                            </a>
                            <a href="https://twitter.com/golangflow" class="btn btn-outline-info">
                                <i class="fab fa-twitter"></i> Twitter
                            </a>
                            <a href="https://github.com/bscott/golangflow" class="btn btn-outline-dark">
                                <i class="fab fa-github"></i> GitHub
                            </a>
                        </div>
                    </div>
                </div>
                <p class="mt-4 text-muted">
                    <small>Go Community Linklog - Powered by <a href="https://gobuffalo.io">Buffalo</a></small>
                </p>
            </div>
        </div>
    </div>
</body>
</html>`
	return c.Render(200, r.String(html))
}

// RSSFeed renders RSS feed
func RSSFeed(c buffalo.Context) error {
	txValue := c.Value("tx")
	if txValue == nil {
		return errors.New("database transaction not available")
	}
	tx := txValue.(*pop.Connection)
	posts := models.Posts{}
	err := tx.Order("created_at desc").All(&posts)
	if err != nil {
		return errors.WithStack(err)
	}

	feed := feeds.Feed{
		Title:       "Golang Flow",
		Link:        &feeds.Link{Href: App().Host},
		Description: "All the Go news that's fit to print!",
		Author:      &feeds.Author{Name: "Brian Scott"},
		Created:     time.Now(),
		Copyright:   "This work is copyright © Brian Scott",
		Items:       make([]*feeds.Item, len(posts), len(posts)),
	}

	for i, p := range posts {
		u := &models.User{}
		err := tx.Find(u, p.UserID)
		if err != nil {
			return errors.WithStack(err)
		}
		feed.Items[i] = &feeds.Item{
			Title:       p.Title,
			Link:        &feeds.Link{Href: fmt.Sprintf("%s/posts/%s", App().Host, p.ID)},
			Description: stripmd.Strip(p.Content),
			Author:      &feeds.Author{Name: u.Name},
			Created:     p.CreatedAt,
		}
	}

	return c.Render(200, r.Func("application/rss+xml", func(w io.Writer, d render.Data) error {
		s, err := feed.ToRss()
		if err != nil {
			return errors.WithStack(err)
		}
		w.Write([]byte(s))
		return nil
	}))
}

//JSONFeed API
func JSONFeed(c buffalo.Context) error {
	txValue := c.Value("tx")
	if txValue == nil {
		return errors.New("database transaction not available")
	}
	tx := txValue.(*pop.Connection)
	posts := models.Posts{}
	err := tx.Order("created_at desc").All(&posts)
	if err != nil {
		return errors.WithStack(err)
	}

	return c.Render(200, r.JSON(posts))
}

//Privacy
func Privacy(c buffalo.Context) error {
	return c.Render(200, r.HTML("templates/privacy.html"))
}

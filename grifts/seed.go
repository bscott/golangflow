package grifts

import (
	"fmt"

	"github.com/bscott/golangflow/models"
	"github.com/markbates/grift/grift"
	"github.com/pkg/errors"
)

var _ = grift.Add("db:seed", func(c *grift.Context) error {
	err := models.DB.TruncateAll()
	u := &models.User{
		Name:           "DB Seed",
		Provider:       "github",
		ProviderUserid: "1234",
	}
	err = models.DB.Create(u)
	if err != nil {
		return errors.WithStack(err)
	}
	if err != nil {
		return errors.WithStack(err)
	}
	for i := 0; i < 500; i++ {
		p := &models.Post{
			Title:   fmt.Sprintf("Post %d", i),
			Content: fmt.Sprintf("content for %d", i),
			UserID:  u.ID,
		}
		err = models.DB.Create(p)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	// Add DB seeding stuff here
	return nil
})

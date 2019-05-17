package registry

import (
	"log"
	"sort"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/nokia/docker-registry-client/registry"
)

// Registry encapsulates the interactions with the registry.
type Registry struct {
	*registry.Registry
}

func init() {
	viper.SetDefault("registry.url", "https://registry-1.docker.io/")
	viper.SetDefault("tag.seperator", "RC")
	viper.SetDefault("tag.dateformat", "20060102")
	viper.SetDefault("tag.keepmin", 10)
	viper.SetDefault("tag.maxage", time.Hour*24*5)
}

// New returns an initialized registry that is ready to use.
func New() (*Registry, error) {
	reg, err := registry.New(
		viper.GetString("registry.url"),
		viper.GetString("registry.username"),
		viper.GetString("registry.password"),
	)
	if err != nil {
		return nil, err
	}
	return &Registry{reg}, nil
}

// GetTags fetches tags from the repository.
func (r *Registry) GetTags(repo string) ([]string, error) {
	return r.Tags(repo)
}

// FindBadTags finds tags that can't be parsed.
func (r *Registry) FindBadTags(tags []string) ([]string, []string, error) {
	badTags := []string{}
	goodTags := []string{}

	// Parse the tags
	for i := range tags {
		// Split on the seperator
		parts := strings.Split(tags[i], viper.GetString("tag.seperator"))
		if len(parts) != 2 {
			badTags = append(badTags, tags[i])
			continue
		}

		// Age must parse as an 8 digit date
		_, err := time.Parse(viper.GetString("tag.dateformat"), parts[0])
		if err != nil {
			badTags = append(badTags, tags[i])
			continue
		}
		goodTags = append(goodTags, tags[i])
	}
	return goodTags, badTags, nil
}

// SortTagsFull sorts the entire tag space, and considers the
// information after the seperator.  This is useful for tags of the
// format <date><seperator><rel> which are sortable.
func (r *Registry) SortTagsFull(tags []string) []string {
	sort.Strings(tags)
	return tags
}

// SortTagsByDate sorts only the date component of the tag.  This is
// generally what is used to compute the number of images to keep for
// things that don't use the RC<rel> format.
func (r *Registry) SortTagsByDate(tags []string) []string {
	sort.Slice(tags, func(i, j int) bool {
		ds1 := strings.Split(tags[i], viper.GetString("tag.seperator"))[0]
		ds2 := strings.Split(tags[j], viper.GetString("tag.seperator"))[0]

		// We perform this unchecked here since this can only
		// be called with well formatted tags.
		d1, _ := time.Parse(viper.GetString("tag.dateformat"), ds1)
		d2, _ := time.Parse(viper.GetString("tag.dateformat"), ds2)

		// Provide the comparison operation
		return d1.Before(d2)
	})
	return tags
}

// KeepTags computes the tags to keep and those to remove.
func (r *Registry) KeepTags(tags []string) ([]string, []string) {
	maxAge := viper.GetDuration("tag.maxage")
	keepAfter := time.Now().Add(maxAge * -1)

	keep := []string{}
	toss := []string{}

	// Pad the date out until the minimum number of tags is kept.
	padDuration := time.Hour * 24 * -1
	for len(keep) < viper.GetInt("tag.keepmin") {
		keepAfter = keepAfter.Add(padDuration)
		log.Println(keepAfter)
		keep, toss = r.keepAfter(keepAfter, tags)
	}
	return keep, toss
}

func (r *Registry) keepAfter(date time.Time, tags []string) ([]string, []string) {
	keep := []string{}
	toss := []string{}

	// Run the split based on the date
	for _, t := range tags {
		parts := strings.Split(t, viper.GetString("tag.seperator"))
		tagdate, _ := time.Parse(viper.GetString("tag.dateformat"), parts[0])
		if tagdate.Before(date) {
			toss = append(toss, t)
			continue
		}
		keep = append(keep, t)
	}

	return keep, toss
}

// RemoveTags is used to actually prune tags.  This does not remove
// layers which may remain referenced, and will be removed during
// garbage collection.
func (r *Registry) RemoveTags(repo string, tags []string) {
	for _, t := range tags {
		digest, err := r.ManifestDigest(repo, t)
		if err != nil {
			log.Println(err)
			continue
		}
		if err := r.DeleteManifest(repo, digest); err != nil {
			log.Println(err)
		}
	}
}

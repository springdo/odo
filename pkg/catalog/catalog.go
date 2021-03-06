package catalog

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/occlient"
	log "github.com/sirupsen/logrus"
)

type CatalogImage struct {
	Name string
	Tags []string
}

// List lists all the available component types
func List(client *occlient.Client) ([]CatalogImage, error) {

	catalogList, err := getDefaultBuilderImages(client)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get image streams")
	}

	if len(catalogList) == 0 {
		return nil, errors.New("unable to retrieve any catalog images from the OpenShift cluster")
	}

	return catalogList, nil
}

// Search searches for the component
func Search(client *occlient.Client, name string) ([]string, error) {
	var result []string
	componentList, err := List(client)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list components")
	}

	// do a partial search in all the components
	for _, component := range componentList {
		if strings.Contains(component.Name, name) {
			result = append(result, component.Name)
		}
	}

	return result, nil
}

// Exists returns true if the given component type is valid, false if not
func Exists(client *occlient.Client, componentType string) (bool, error) {
	catalogList, err := List(client)
	if err != nil {
		return false, errors.Wrapf(err, "unable to list catalog")
	}

	for _, supported := range catalogList {
		if componentType == supported.Name {
			return true, nil
		}
	}
	return false, nil
}

// getDefaultBuilderImages returns the default builder images available in the
// openshift namespace
func getDefaultBuilderImages(client *occlient.Client) ([]CatalogImage, error) {
	imageStreams, err := client.GetImageStreams(occlient.OpenShiftNameSpace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get Image Streams")
	}

	var builderImages []CatalogImage

	// Get builder images from the available imagestreams
	for _, imageStream := range imageStreams {
		var allTags []string
		buildImage := false

		for _, tag := range imageStream.Spec.Tags {

			allTags = append(allTags, tag.Name)

			// Check to see if it is a "builder" image
			if _, ok := tag.Annotations["tags"]; ok {
				for _, t := range strings.Split(tag.Annotations["tags"], ",") {

					// If the tag has "builder" then we will add the image to the list
					if t == "builder" {
						buildImage = true
					}
				}
			}

		}

		// Append to the list of images if a "builder" tag was found
		if buildImage {
			builderImages = append(builderImages, CatalogImage{Name: imageStream.Name, Tags: allTags})
		}

	}

	log.Debugf("Found builder images: %v", builderImages)
	return builderImages, nil
}

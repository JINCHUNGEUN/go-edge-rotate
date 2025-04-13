package main

import (
	"flag"

	rotate "github.com/JINCHUNGEUN/go-edge-rotate/pkg/rotate"
)

func main() {
	var (
		imagePath  string
		tileSize   int
		angle      float64
		action     string
		outputDir  string
		ratio      float64
		mapName    string
		mapMd5     string
		sourceData string
		newData    string
		parentDir  string
	)
	const (
		ACTION_ROTATE_IMAGE = "rotateImage"
		ACTION_SPLIT_IMAGE  = "splitImage"
		ACTION_MERGE_IMAGE  = "mergeImage"
	)

	flag.StringVar(&action, "action", "default", "action")
	flag.StringVar(&imagePath, "imagePath", "default", "image path stripping suffix")
	flag.IntVar(&tileSize, "tileSize", 100, "block size")
	flag.Float64Var(&angle, "angle", 0.0, "rotate angle")
	flag.Float64Var(&ratio, "ratio", 0.0, "ratio")
	flag.StringVar(&outputDir, "outputDir", "", "")
	flag.StringVar(&mapName, "mapName", "RoboGo-1-1-1", "")
	flag.StringVar(&mapMd5, "mapMd5", "", "")
	flag.StringVar(&sourceData, "sourceData", "", "")
	flag.StringVar(&newData, "newData", "", "")
	flag.StringVar(&parentDir, "parentDir", "", "")
	flag.Parse()

	switch action {
	case ACTION_ROTATE_IMAGE:
		rotate.RotateImage(imagePath, tileSize, angle)
	case ACTION_SPLIT_IMAGE:
		rotate.SplitImage(mapName, imagePath, mapMd5, tileSize, outputDir, ratio, parentDir)
	case ACTION_MERGE_IMAGE:
		rotate.MergeImage(sourceData, newData, imagePath)
	}
}

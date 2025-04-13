package rotate

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gocv.io/x/gocv"
)

func GetRotatedImageSize(angle float64, original image.Point) image.Point {
	radian := angle * math.Pi / 180
	w := original.X
	h := original.Y
	cos := math.Cos(radian)
	sin := math.Sin(radian)
	W := int(math.Abs(float64(w)*cos) + math.Abs(float64(h)*sin))
	H := int(math.Abs(float64(w)*sin) + math.Abs(float64(h)*cos))
	return image.Point{X: W, Y: H}
}

func RotateImage(imagePath string, tileSize int, angle float64) {
	srcFilePath := imagePath + ".png"

	img := gocv.IMRead(srcFilePath, gocv.IMReadGrayScale)
	center := image.Point{img.Cols() / 2, img.Rows() / 2}
	if img.Empty() {
		fmt.Println("Error reading image")
		return
	}
	srcWidth := img.Cols()
	srcHeight := img.Rows()

	newImageSize := GetRotatedImageSize(angle, image.Point{img.Cols(), img.Rows()})
	newImg := gocv.NewMatWithSize(newImageSize.Y, newImageSize.X, img.Type())

	fmt.Println("original image size: ", img.Cols(), " ", img.Rows())
	fmt.Println("new image size: ", newImg.Cols(), " ", newImg.Rows())
	backgroundColor := gocv.NewScalar(128, 128, 128, 1)
	newImg.SetTo(backgroundColor)

	scale := 1.0 // 缩放比例为 1

	rows := (srcHeight + tileSize - 1) / tileSize
	cols := (srcWidth + tileSize - 1) / tileSize
	fmt.Println("new image size: ", rows, " ", cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			w := tileSize
			if w+2 > srcWidth-j*tileSize {
				w = srcWidth - j*tileSize - 2
			}
			h := tileSize
			if h+2 > srcHeight-i*tileSize {
				h = srcHeight - i*tileSize - 2
			}
			tile := img.Region(image.Rect(j*tileSize, i*tileSize, j*tileSize+w+2, i*tileSize+h+2))
			// 将原始图像转换为具有 Alpha 通道的图像（RGBA）
			var srcImgBGRA = gocv.NewMat()
			tileWidth := tile.Cols()
			tileHeight := tile.Rows()
			gocv.CvtColor(tile, &srcImgBGRA, gocv.ColorBGRToBGRA)
			alpha := gocv.NewMatWithSize(tileHeight, tileWidth, gocv.MatTypeCV8U)
			alpha.SetTo(gocv.Scalar{Val1: 255})
			gocv.InsertChannel(alpha, &srcImgBGRA, 3)

			centerX := float32(tileWidth) / 2
			centerY := float32(tileHeight) / 2
			rotationMatrix := gocv.GetRotationMatrix2D(image.Point{int(centerX), int(centerY)}, -angle, scale)

			cosVal := math.Abs(float64(rotationMatrix.GetDoubleAt(0, 0)))
			sinVal := math.Abs(float64(rotationMatrix.GetDoubleAt(0, 1)))

			newCols := int(float64(tileWidth)*cosVal + float64(tileHeight)*sinVal)
			newRows := int(float64(tileWidth)*sinVal + float64(tileHeight)*cosVal)
			rotationMatrix.SetDoubleAt(0, 2, rotationMatrix.GetDoubleAt(0, 2)+float64(newCols)/2-float64(centerX))
			rotationMatrix.SetDoubleAt(1, 2, rotationMatrix.GetDoubleAt(1, 2)+float64(newRows)/2-float64(centerY))
			var rotatedTile = gocv.NewMatWithSize(newRows, newCols, srcImgBGRA.Type())
			gocv.WarpAffineWithParams(
				srcImgBGRA, &rotatedTile, rotationMatrix, image.Point{X: newCols, Y: newRows}, gocv.InterpolationLinear, gocv.BorderTransparent, color.RGBA{0, 0, 0, 0},
			)
			rotatedBlockW := rotatedTile.Cols()
			rotatedBlockH := rotatedTile.Rows()
			tile.Close()
			rotationMatrix.Close()
			srcImgBGRA.Close()
			alpha.Close()

			var smallImageAlpha = gocv.NewMat()
			gocv.ExtractChannel(rotatedTile, &smallImageAlpha, 3)
			var smallGrayImage = gocv.NewMat()
			gocv.CvtColor(rotatedTile, &smallGrayImage, gocv.ColorBGRAToGray)
			rotatedTile.Close()

			var (
				tileCenterX int
				tileCenterY int
				tileX       int
				tileY       int
			)
			tileX = j * tileSize
			if tileWidth > tileSize {
				tileCenterX = tileX + tileSize/2
			} else {
				tileCenterX = tileX + tileWidth/2 - 1
			}
			tileY = i * tileSize
			if tileHeight > tileSize {
				tileCenterY = tileY + tileSize/2
			} else {
				tileCenterY = tileY + tileHeight/2 - 1
			}
			offsetX := tileCenterX - center.X
			offsetY := tileCenterY - center.Y

			largeCenterX := float64(newImg.Cols()) / 2
			largeCenterY := float64(newImg.Rows()) / 2

			radianAngle := angle * math.Pi / 180
			rotatedCenterX := int(largeCenterX + float64(offsetX)*math.Cos(radianAngle) - float64(offsetY)*math.Sin(radianAngle))
			rotatedCenterY := int(largeCenterY + float64(offsetX)*math.Sin(radianAngle) + float64(offsetY)*math.Cos(radianAngle))

			rotatedX := rotatedCenterX - rotatedBlockW/2
			rotatedY := rotatedCenterY - rotatedBlockH/2
			if newImg.Cols()-rotatedBlockW < rotatedX {
				rotatedX = newImg.Cols() - rotatedBlockW
			}
			if rotatedX < 0 {
				rotatedX = 0
			}
			if newImg.Rows()-rotatedBlockH < rotatedY {
				rotatedY = newImg.Rows() - rotatedBlockH
			}
			if rotatedY < 0 {
				rotatedY = 0
			}

			zeroMask := gocv.NewMatWithSize(smallImageAlpha.Rows(), smallImageAlpha.Cols(), gocv.MatTypeCV8U)
			gocv.Compare(smallImageAlpha, zeroMask, &zeroMask, gocv.CompareEQ)
			smallGrayImage.CopyToWithMask(&smallGrayImage, zeroMask)
			zeroMask.Close()
			smalGrayImageWidth := smallGrayImage.Cols()
			smalGrayImageHeight := smallGrayImage.Rows()
			mask := gocv.NewMatWithSize(smalGrayImageHeight, smalGrayImageWidth, gocv.MatTypeCV8U)
			gocv.Compare(smallImageAlpha, mask, &mask, gocv.CompareGT)
			smallImageAlpha.Close()

			newTile := newImg.Region(image.Rect(rotatedX, rotatedY, rotatedX+rotatedBlockW, rotatedY+rotatedBlockH))
			smallGrayImage.CopyToWithMask(&newTile, mask)
			mask.Close()
			newTile.Close()
			smallGrayImage.Close()
		}
	}
	img.Close()
	outputFile := imagePath + ".save.png"
	gocv.IMWrite(outputFile, newImg)
	newImg.Close()
}

type ImageSliceInfo struct {
	X      int
	Y      int
	Width  int
	Height int
	Path   string
}

func SplitImageBySize(imagePath string, size int, outputDir string) []ImageSliceInfo {
	images := []ImageSliceInfo{}
	os.MkdirAll(outputDir, os.ModePerm)
	img := gocv.IMRead(imagePath, gocv.IMReadUnchanged)
	if img.Empty() {
		fmt.Println("Error reading image 2")
		return images
	}
	width := img.Cols()
	height := img.Rows()
	cols := (width + size - 1) / size
	rows := (height + size - 1) / size
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			w := size
			if w > width-j*size {
				w = width - j*size
			}
			h := size
			if h > height-i*size {
				h = height - i*size
			}
			tile := img.Region(image.Rect(j*size, i*size, j*size+w, i*size+h))
			fileName := strconv.Itoa(j) + "_" + strconv.Itoa(i) + ".png"
			finalPath := filepath.Join(outputDir, fileName)
			gocv.IMWrite(finalPath, tile)
			tile.Close()
			images = append(images, ImageSliceInfo{j * size, i * size, w, h, finalPath})
		}
	}
	img.Close()
	return images
}

func MergeImage(sourceImageData string, newImageData string, imagePath string) {

	newImageDataDir := filepath.Dir(newImageData)
	subImages := SplitImageBySize(newImageData, 5000, filepath.Join(newImageDataDir, "subImages"))

	img1 := gocv.IMRead(sourceImageData, gocv.IMReadGrayScale)
	if img1.Empty() {
		fmt.Println("Error reading image 1")
		return
	}
	defer img1.Close()

	for idx := range subImages {
		img2 := gocv.IMRead(subImages[idx].Path, gocv.IMReadUnchanged)
		if img2.Empty() {
			fmt.Println("Error reading image 2")
			return
		}
		offsetX := subImages[idx].X
		offsetY := subImages[idx].Y
		blockSize := 2000
		rows, cols := img2.Rows(), img2.Cols()
		rowBlocks, colBlocks := (rows+blockSize-1)/blockSize, (cols+blockSize-1)/blockSize
		for i := 0; i < rowBlocks; i++ {
			for j := 0; j < colBlocks; j++ {
				// 获取 img1 和 img2 的小块区域
				x, y := j*blockSize, i*blockSize
				width, height := blockSize, blockSize
				if x+width > cols {
					width = cols - x
				}
				if y+height > rows {
					height = rows - y
				}

				roi2 := img2.Region(image.Rect(x, y, x+width, y+height))
				channels := gocv.Split(roi2)
				if len(channels) < 4 {
					roi2.Close()
					for idx := range channels {
						channels[idx].Close()
					}
					continue
				}
				gocv.CvtColor(roi2, &roi2, gocv.ColorRGBAToGray)
				minVal, maxVal, _, _ := gocv.MinMaxLoc(channels[3])
				if minVal == 0 && maxVal == 0 {
					roi2.Close()
					for idx := range channels {
						channels[idx].Close()
					}
					continue
				}
				channels[0].Close()
				channels[1].Close()
				channels[2].Close()
				roi1 := img1.Region(image.Rect(offsetX+x, offsetY+y, offsetX+x+width, offsetY+y+height))

				mask := channels[3]

				// 通过像素值范围，将非白非黑区域设置颜色值为128
				maskNew := gocv.NewMat()
				gocv.InRangeWithScalar(roi2, gocv.NewScalar(100, 100, 100, 100), gocv.NewScalar(200, 200, 200, 200), &maskNew)
				newRoi2 := gocv.NewMat()
				roi2.CopyTo(&newRoi2)
				maxValMat := gocv.NewMatWithSize(roi2.Rows(), roi2.Cols(), roi2.Type())
				maxValMat.SetTo(gocv.NewScalar(128, 0, 0, 0))
				roi2.Close()
				maxValMat.CopyToWithMask(&newRoi2, maskNew)
				maxValMat.Close()
				maskNew.Close()

				invertedMask := gocv.NewMatWithSize(height, width, roi1.Type())
				gocv.BitwiseNot(mask, &invertedMask)

				maskB := gocv.NewMat()
				gocv.Threshold(invertedMask, &maskB, 254, 255, gocv.ThresholdBinary)
				backMask := gocv.NewMat()
				invertedMask.CopyToWithMask(&backMask, maskB)

				foreground := gocv.NewMatWithSize(height, width, roi1.Type())
				newRoi2.CopyToWithMask(&foreground, mask)
				newRoi2.Close()
				mask.Close()

				background := gocv.NewMatWithSize(height, width, roi1.Type())
				roi1.CopyToWithMask(&background, backMask)
				backMask.Close()
				invertedMask.Close()

				merged := gocv.NewMatWithSize(height, width, roi1.Type())
				gocv.Add(foreground, background, &merged)
				foreground.Close()
				background.Close()
				// 将合并的小块复制回原始图像的 ROI
				merged.CopyTo(&roi1)
				merged.Close()
				roi1.Close()
			}
		}

		img2.Close()
	}

	gocv.IMWrite(imagePath, img1)
}

func MergeImageV2(image1Filename, image2Filename, imagePath string) {
	// 读取图像
	img1File, err := os.Open(image1Filename)
	if err != nil {
		fmt.Println("Error reading image 1:", err)
		return
	}

	img1, err := png.Decode(img1File)
	img1File.Close()
	if err != nil {
		fmt.Println("Error decoding image 1:", err)
		return
	}

	// 创建一个新的图像，它与第一个图像具有相同的大小
	newImg := image.NewRGBA(img1.Bounds())

	// 将第一个图像的像素复制到新图像上
	draw.Draw(newImg, newImg.Bounds(), img1, image.Point{}, draw.Src)

	img2File, err := os.Open(image2Filename)
	if err != nil {
		fmt.Println("Error reading image 2:", err)
		return
	}
	img2, err := png.Decode(img2File)
	img2File.Close()
	if err != nil {
		fmt.Println("Error decoding image 2:", err)
		return
	}

	// 将第二个图像的非透明部分复制到新图像上
	draw.Draw(newImg, newImg.Bounds(), img2, image.Point{}, draw.Over)

	outFile, err := os.Create(imagePath)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outFile.Close()

	err = png.Encode(outFile, newImg)
	if err != nil {
		fmt.Println("Error encoding merged image:", err)
		return
	}
	fmt.Println("Merged image saved as", imagePath)
}

func base64DecodeImage(base64Image string) ([]byte, error) {
	dataURLPrefix := "data:image/png;base64,"
	if strings.HasPrefix(base64Image, dataURLPrefix) {
		base64Image = strings.TrimPrefix(base64Image, dataURLPrefix)
	}

	return base64.StdEncoding.DecodeString(base64Image)
}

func SplitImage(mapName string, filePath string, mapMd5 string, tileSize int, parentDir string, ratio float64, executeDir string) {
	fmt.Println("start split image: ", filePath, " tilesize: ", tileSize, " outputDir: ", parentDir, " ratio: ", ratio)
	img := gocv.IMRead(filePath, gocv.IMReadGrayScale)
	srcImgWidth := img.Cols()
	srcImgHeight := img.Rows()
	currentRatio := 1.0
	currentImgWidth := srcImgWidth
	currentImgHeight := srcImgHeight
	loopCount := 0
	outputDir := filepath.Join(parentDir, mapName)
	os.MkdirAll(outputDir, os.ModePerm)
	metaInfoPath := filepath.Join(outputDir, "info.json")
	metaInfoSlices := []map[string]any{}
	mapThumbPath := ""
	var mapThumbWidth, mapThumbHeight int
	thumbSize := 1000
	for {
		if srcImgWidth < tileSize && srcImgHeight < tileSize {
			fmt.Println("current image is small enough break")
			break
		}
		loopCount += 1
		var resizedImg gocv.Mat
		if loopCount == 1 {
			resizedImg = img.Clone()
		} else {
			currentRatio = math.Pow(ratio, float64(loopCount-1))
			currentImgWidth = int(float64(srcImgWidth) * currentRatio)
			currentImgHeight = int(float64(srcImgHeight) * currentRatio)
			resizedImg = gocv.NewMatWithSize(currentImgHeight, currentImgWidth, img.Type())
			fmt.Println("cur loop: ", loopCount, " ", currentRatio, " ", currentImgWidth, " ", currentImgHeight, " ", resizedImg.Channels())
			if currentImgHeight == 0 || currentImgWidth == 0 {
				break
			}
			gocv.Resize(img, &resizedImg, image.Point{X: currentImgWidth, Y: currentImgHeight}, 0, 0, gocv.InterpolationLinear)
			if currentImgHeight <= tileSize && currentImgWidth <= tileSize {
				fileName := strconv.Itoa(loopCount) + "_0_0.png"
				sliceOutputPath := filepath.Join(outputDir, fileName)
				gocv.IMWrite(sliceOutputPath, resizedImg)
				resizedImg.Close()
				metaInfoSlices = append(
					metaInfoSlices, map[string]any{
						"path": strings.ReplaceAll(sliceOutputPath, executeDir, ""), "x": 0, "y": 0, "resolution": currentRatio,
						"tileSize": tileSize,
					},
				)
				break
			}
		}
		if currentImgWidth <= thumbSize || currentImgHeight <= thumbSize {
			if mapThumbPath == "" {
				mapThumbPath = filepath.Join(outputDir, "thumb.png")
				gocv.IMWrite(mapThumbPath, resizedImg)
				mapThumbPath = strings.ReplaceAll(mapThumbPath, executeDir, "")
				mapThumbHeight = currentImgHeight
				mapThumbWidth = currentImgWidth
			}
		}
		cols := (currentImgWidth + tileSize - 1) / tileSize
		rows := (currentImgHeight + tileSize - 1) / tileSize
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				w := tileSize
				if w > currentImgWidth-j*tileSize {
					w = currentImgWidth - j*tileSize
				}
				h := tileSize
				if h > currentImgHeight-i*tileSize {
					h = currentImgHeight - i*tileSize
				}
				tile := resizedImg.Region(image.Rect(j*tileSize, i*tileSize, j*tileSize+w, i*tileSize+h))
				fmt.Println("tile size: ", tile.Cols(), " ", tile.Rows())
				fileName := strconv.Itoa(loopCount) + "_" + strconv.Itoa(i) + "_" + strconv.Itoa(j) + ".png"
				sliceOutputPath := filepath.Join(outputDir, fileName)
				gocv.IMWrite(sliceOutputPath, tile)
				tile.Close()
				metaInfoSlices = append(
					metaInfoSlices, map[string]any{
						"path": strings.ReplaceAll(sliceOutputPath, executeDir, ""), "x": j, "y": i, "resolution": currentRatio,
						"tileSize": tileSize,
					},
				)
			}
		}
		resizedImg.Close()
	}
	img.Close()

	fp, err := os.Create(metaInfoPath)
	if err != nil {
		fmt.Println("create file failed. ", err)
		return
	}
	defer fp.Close()
	metaInfo := map[string]any{
		"mapName": mapName, "tiles": metaInfoSlices,
		"mapMd5":        mapMd5,
		"mapThumbPath":  mapThumbPath,
		"mapThumbWidth": mapThumbWidth, "mapThumbHeight": mapThumbHeight,
	}
	data, err := json.Marshal(metaInfo)
	if err != nil {
		fmt.Println("marshal meta info failed. ", err)
		return
	}
	_, err = fp.Write(data)
	if err != nil {
		fmt.Println("write meta info to file failed. ", err)
	}
}

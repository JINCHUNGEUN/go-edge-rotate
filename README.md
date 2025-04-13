# GoCV Image Utilities README Q&A

## What does this package do?
This package provides image manipulation utilities for rotating, splitting, and merging images. It's particularly designed for working with large map images that need to be processed in tiles for efficiency.

## What are the key features?
- Image rotation with tile-based processing for large images
- Multi-resolution image splitting (pyramid-like) for map tiling
- Image merging with transparency handling
- Support for generating metadata for tiled maps

## What are the dependencies?
- [GoCV](https://gocv.io/) - Go bindings for OpenCV
- Standard Go image libraries

## What is the suitable application scenario?
This package is particularly well-suited for edge computing applications with limited resources. The tile-based processing approach allows handling large images efficiently on devices with constrained memory and processing power. This makes it ideal for:
- IoT image processing devices
- Edge servers with limited resources
- Mobile applications that need to process large images
- Embedded systems with image processing requirements

## How do I use the RotateImage function?
```bash
go run main.go -action rotateImage -imagePath path/to/image -tileSize 500 -angle 45
```
This will rotate the image at `path/to/image.png` by 45 degrees, processing it in tiles of 500x500 pixels. The output will be saved as `path/to/image.save.png`.

## How do I split an image for multi-resolution viewing?
```bash
go run main.go -action splitImage -imagePath path/to/image.png -mapName MyMap -tileSize 256 -outputDir ./output -ratio 0.5 -parentDir ./current/dir
```
This splits the image into tiles at multiple resolution levels, with each level at half the size of the previous one. The output includes a `info.json` file with metadata about all the tiles.

## How do I merge two images?
```bash
go run main.go -action mergeImage -sourceData path/to/base.png -newData path/to/overlay.png -imagePath path/to/output.png
```
This will merge the overlay image on top of the base image, respecting transparency, and save the result to the output path.

## What are the command line options?
- `-action`: Specifies the operation to perform (`rotateImage`, `splitImage`, or `mergeImage`)
- `-imagePath`: Path to the input image (for rotation and splitting) or output image (for merging)
- `-tileSize`: Size of tiles for processing (default: 100)
- `-angle`: Rotation angle in degrees (for rotation)
- `-ratio`: Scale ratio between consecutive resolution levels (for splitting)
- `-outputDir`: Directory to store output tiles (for splitting)
- `-mapName`: Name of the map (for splitting)
- `-mapMd5`: MD5 hash of the map (for splitting)
- `-sourceData`: Path to the base image (for merging)
- `-newData`: Path to the overlay image (for merging)
- `-parentDir`: Parent directory for relative path calculation (for splitting)

## Detailed Analysis of Core Methods

### GetRotatedImageSize Method
This method calculates the dimensions of an image after rotation:
- Accepts rotation angle and original image dimensions as input
- Converts the angle to radians
- Calculates the rotated width and height using mathematical formulas
- The dimensions are based on trigonometric functions (cosine and sine) to ensure coverage of all image content

### RotateImage Method
This is a complex method that efficiently rotates large images through tile-based processing:

1. **Tile-based Processing**:
   - Divides the original image into multiple small blocks (tiles) for processing, rather than processing the entire image at once
   - This significantly reduces memory usage and avoids memory overflow issues when working with large images

2. **Processing Flow**:
   - Creates a new, larger image to contain the rotated image
   - Divides the image into blocks according to the specified tileSize
   - Rotates each block individually
   - Calculates the position of each rotated block in the new image
   - Uses masks to handle semi-transparent areas

3. **Optimization Points**:
   - Promptly closes Mat objects that are no longer needed to free up memory
   - Only operates on specific regions, avoiding processing the entire image
   - Uses the Region method to get image sub-regions, avoiding copying large blocks of data

### SplitImageBySize Method
Splits an image into fixed-size blocks:
- Creates the output directory
- Divides the image into a grid of the specified size
- Gets the corresponding image region for each grid cell
- Saves each region as a separate PNG file
- Returns an array containing information about all slices (position, size, path)
- Uses the Region method to reference areas of the original image, avoiding extensive memory copying

### MergeImage Method
Merges two images while respecting transparency information:

1. **Chunking Strategy**:
   - First divides the new image into sub-images of size 5000
   - Then divides these sub-images into smaller 2000×2000 blocks for processing
   - This multi-level chunking strategy can handle very large images without exhausting memory

2. **Merging Logic**:
   - Extracts the alpha channel of each block as a mask
   - Distinguishes between foreground and background based on the mask
   - Only performs image merging where there are alpha channel values
   - Special handling for middle gray areas (values between 100-200)

3. **Resource Optimization**:
   - Processes block by block rather than the entire image
   - Promptly closes temporarily created Mat objects
   - Skips processing of completely transparent areas

### MergeImageV2 Method
This is a simplified version of the merge method implemented using Go's standard library:
- Uses the standard image package and draw package
- Less memory efficient but simpler implementation
- Suitable for quick merging of smaller images
- Directly uses Go's Draw method to handle transparency

### SplitImage Method
Creates a multi-resolution image pyramid suitable for map applications:

1. **Multi-resolution Processing**:
   - Starts with the original image
   - Gradually reduces the image by the specified ratio
   - Creates image blocks at each size level
   - Continues until the image is smaller than a single tile size

2. **Metadata Generation**:
   - Records information about each image block (path, position, resolution)
   - Generates a thumbnail
   - Saves all information to a JSON file

3. **Resource Optimization**:
   - Processes each resolution level in a loop
   - Closes related resources when the current level processing is complete
   - Special handling for smaller sizes (less than a single tile)

## Overall Resource Optimization Strategy

1. **Chunked Processing**: All complex operations use a chunking strategy to avoid loading large amounts of data into memory at once

2. **Resource Release**: Uses defer and manual Close() methods to ensure timely release of resources

3. **Reference Instead of Copy**: Uses the Region method to get image sub-regions, avoiding unnecessary data copying

4. **Parallel Processing Potential**: Code structure allows for easy extension to parallel processing (although the current implementation is serial)

5. **Memory Control**:
   - Carefully manages OpenCV's Mat objects (these are wrappers around underlying resources)
   - Uses appropriate image types (such as grayscale images to reduce memory usage)
   - Promptly releases temporary objects that are no longer needed

These methods, used in combination, can efficiently handle large image operations, making them particularly suitable for map data processing and multi-resolution image management scenarios, especially in edge computing environments with limited resources.

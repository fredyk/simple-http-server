package simple_http_server

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	wst "github.com/fredyk/westack-go/v2/common"
	"github.com/fredyk/westack-go/v2/model"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var PHYSICAL_ASSET_PATH = "./assets/dist/"
var port = 8080
var _debug = false

type Options struct {
	PhysicalAssetPath string
	Port              int
	Debug             bool
}

func InitAndServe(options ...Options) error {
	if len(options) > 0 {
		if options[0].PhysicalAssetPath != "" {
			PHYSICAL_ASSET_PATH = options[0].PhysicalAssetPath
		}
		if options[0].Port > 0 {
			port = options[0].Port
		}
		_debug = options[0].Debug
	} else {
		if os.Getenv("PHYSICAL_ASSET_PATH") != "" {
			PHYSICAL_ASSET_PATH = os.Getenv("PHYSICAL_ASSET_PATH")
		}
		if os.Getenv("PORT") != "" {
			port, _ = strconv.Atoi(os.Getenv("PORT"))
		}
		if os.Getenv("DEBUG") == "true" || os.Getenv("DEBUG") == "1" {
			_debug = true
		}
	}

	if PHYSICAL_ASSET_PATH != "" && PHYSICAL_ASSET_PATH != "/" && PHYSICAL_ASSET_PATH != "." && !strings.HasSuffix(PHYSICAL_ASSET_PATH, "/") {
		PHYSICAL_ASSET_PATH = fmt.Sprintf("%s/", PHYSICAL_ASSET_PATH)
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: !_debug,
	})
	app.Use(recover.New(recover.Config{
		StackTraceHandler: func(c *fiber.Ctx, err interface{}) {
			fmt.Printf("Error: %s\n", err)
			debug.PrintStack()
			err = c.Status(500).JSON(wst.M{"error": fmt.Sprintf("%s", err)})
			if err != nil {
				fmt.Printf("Error sending error response: %s\n", err)
				os.Exit(1)
			}
		},
		EnableStackTrace: true,
	}))

	// disable CORs
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, HEAD")
		// all headers
		c.Set("Access-Control-Allow-Headers", "*")
		return c.Next()
	})

	var asset404 interface{}
	app.Use(func(c *fiber.Ctx) error {
		ctx := &model.EventContext{

			Data: &wst.M{
				"path":      c.Path(),
				"base_path": "/",
			},
		}
		err := SendStaticAsset(ctx, c)
		fileSize := (*ctx.Ephemeral)["fileSize"].(int64)
		if err != nil {
			if os.IsNotExist(err) {
				ctx.StatusCode = 404
				if asset404 != nil {
					ctx.Result = asset404
				} else {
					fileLocation := fmt.Sprintf("%s404.html", PHYSICAL_ASSET_PATH)
					asset404, fileSize, err = readFileBytes(ctx, c, fileLocation)
					if err != nil {
						fileLocation := fmt.Sprintf("%serrors/404.html", PHYSICAL_ASSET_PATH)
						asset404, fileSize, err = readFileBytes(ctx, c, fileLocation)
						if err != nil {
							fmt.Printf("Error reading 404.html: %s\n", err)
							asset404 = []byte("404 Not Found")
							fileSize = int64(len(asset404.([]byte)))
						} else {
							ctx.Result = asset404
						}
					} else {
						ctx.Result = asset404
					}
				}
			} else {
				fmt.Printf("Error reading file: %s\n", err)
				ctx.StatusCode = 500
				ctx.Result = []byte("500 Internal Server Error")
				fileSize = int64(len(ctx.Result.([]byte)))
			}
		}
		return sendResult(ctx, fileSize, c)
	})

	return app.Listen(fmt.Sprintf(":%d", port))
}

func sendResult(ctx *model.EventContext, fileSize int64, c *fiber.Ctx) error {
	// if ctx.Result != nil {
	c.Status(ctx.StatusCode)
	c.Set("Content-Length", fmt.Sprintf("%d", fileSize))
	path := ctx.Data.GetString("path")
	var mimeType string
	var isCss bool
	var isJs bool
	var isMedia bool
	var isFont bool
	var isDocument bool
	if strings.HasSuffix(path, ".css") {
		mimeType = "text/css"
		isCss = true
	} else if strings.HasSuffix(path, ".js") {
		mimeType = "application/javascript"
		isJs = true
	} else if strings.HasSuffix(path, ".json") {
		mimeType = "application/json"
		isDocument = true
	} else if strings.HasSuffix(path, ".html") {
		mimeType = "text/html"
		isDocument = true
	} else if strings.HasSuffix(path, ".xml") {
		mimeType = "application/xml"
		isDocument = true
	} else if strings.HasSuffix(path, ".txt") {
		mimeType = "text/plain"
		isDocument = true
	} else if strings.HasSuffix(path, ".png") {
		mimeType = "image/png"
		isMedia = true
	} else if strings.HasSuffix(path, ".jpg") {
		mimeType = "image/jpeg"
		isMedia = true
	} else if strings.HasSuffix(path, ".gif") {
		mimeType = "image/gif"
		isMedia = true
	} else if strings.HasSuffix(path, ".svg") {
		mimeType = "image/svg+xml"
		isMedia = true
	} else if strings.HasSuffix(path, ".ico") {
		mimeType = "image/x-icon"
		isMedia = true
	} else if strings.HasSuffix(path, ".webp") {
		mimeType = "image/webp"
		isMedia = true
	} else if strings.HasSuffix(path, ".woff") {
		mimeType = "font/woff"
		isFont = true
	} else if strings.HasSuffix(path, ".woff2") {
		mimeType = "font/woff2"
		isFont = true
	} else if strings.HasSuffix(path, ".ttf") {
		mimeType = "font/ttf"
		isFont = true
	} else if strings.HasSuffix(path, ".otf") {
		mimeType = "font/otf"
		isFont = true
	} else if strings.HasSuffix(path, ".eot") {
		mimeType = "font/eot"
		isFont = true
	} else if strings.HasSuffix(path, ".mp4") {
		mimeType = "video/mp4"
		isMedia = true
	} else if strings.HasSuffix(path, ".webm") {
		mimeType = "video/webm"
		isMedia = true
	} else if strings.HasSuffix(path, ".ogg") {
		mimeType = "video/ogg"
		isMedia = true
	} else if strings.HasSuffix(path, ".mp3") {
		mimeType = "audio/mp3"
		isMedia = true
	} else if strings.HasSuffix(path, ".webmanifest") {
		mimeType = "application/manifest+json"
		isDocument = true
	} else {
		// mimeType = "application/octet-stream"
		mimeType = "text/html"
		isDocument = true
	}
	if ctx.StatusCode == 200 && (isCss || isJs || isMedia || isFont || isDocument) {
		var cacheControl string
		fileName := ctx.Data.GetString("path")
		splt := strings.Split(fileName, "/")
		fileName = splt[len(splt)-1]
		isHashed := regexp.MustCompile(`\.[a-f0-9]{6,}\.`).MatchString(fileName)
		if isHashed {
			cacheControl = "max-age=31536000"
		} else {
			if isCss || isJs {
				// 30 minutes
				cacheControl = "max-age=1800"
			} else if isMedia {
				// 1 day
				cacheControl = "max-age=86400"
			} else if isFont {
				// 1 year
				cacheControl = "max-age=31536000"
			} else if isDocument {
				// 1 day
				cacheControl = "max-age=86400"
			}
		}
		if cacheControl != "" {
			if _debug {
				fmt.Printf("[DEBUG] Setting cache control for %s --> %s\n", fileName, cacheControl)
			}
			c.Set("Cache-Control", cacheControl)
		}
	}
	if ctx.Result == nil {
		return c.Send(nil)
	} else if v, ok := ctx.Result.([]byte); ok {
		c.Set("Content-Type", mimeType)
		return c.Send(v)
	} else if v, ok := ctx.Result.(string); ok {
		c.Set("Content-Type", "text/plain")
		return c.SendString(v)
	} else if v, ok := ctx.Result.(wst.M); ok {
		return c.JSON(v)
	} else if v, ok := ctx.Result.(wst.A); ok {
		return c.JSON(v)
	} else if v, ok := ctx.Result.(int); ok {
		return c.SendStatus(v)
	} else /* chan */ if v, ok := ctx.Result.(io.Reader); ok {
		if _debug {
			fmt.Printf("[DEBUG] Sending stream\n")
		}
		c.Set("Content-Type", mimeType)
		return c.SendStream(v)
	} else {
		return c.JSON(ctx.Result)
	}
	// }
	// return c.Status(fiber.ErrBadRequest.Code).SendString(fiber.ErrBadRequest.Message)
}

func convertHttpPathToFileLocation(basePath string, path string) string {
	// Escape path
	// Map path to static file
	// Return file location
	path = strings.ReplaceAll(path, "..", "")

	re := regexp.MustCompile(`\:([a-zA-Z0-9_-]+)`).ReplaceAllString(basePath, "([^/]+)")
	re = strings.ReplaceAll(re, "/*", "")

	re = fmt.Sprintf("^%s/?", re)
	if _debug {
		fmt.Printf("[DEBUG] Regexp: %s\n", re)
	}
	path = regexp.MustCompile(re).ReplaceAllString(path, PHYSICAL_ASSET_PATH)

	return path
}

func readFileBytes(ctx *model.EventContext, c *fiber.Ctx, fileLocation string) (interface{}, int64, error) {
	var fileContent []byte
	var baseName string
	var fileSize int64
	if fileLocation == PHYSICAL_ASSET_PATH {
		fileLocation = fmt.Sprintf("%sindex.html", PHYSICAL_ASSET_PATH)
		if _debug {
			fmt.Printf("New File location: %s\n", fileLocation)
		}
	} else {
		// remove trailing slash
		cleanFileLocation := regexp.MustCompile(`/$`).ReplaceAllString(fileLocation, "")
		splt := strings.Split(cleanFileLocation, "/")
		if len(splt) > 1 {
			baseName = splt[len(splt)-1]
		}
		var extension string
		splt = strings.Split(baseName, ".")
		if len(splt) > 1 {
			extension = splt[len(splt)-1]
		}

		if extension == "" {
			fileLocation = fmt.Sprintf("%s/index.html", cleanFileLocation)
			if _debug {
				fmt.Printf("New File location: %s\n", fileLocation)
			}
		} else {
			if _debug {
				fmt.Printf("Extension: %s\n", extension)
			}
		}
	}
	// if file not exists
	fStat, err := os.Stat(fileLocation)
	if os.IsNotExist(err) {
		return nil, 0, err
	}
	fileSize = fStat.Size()
	if _debug {
		fmt.Printf("File size: %d\n", fileSize)
	}
	f, err := os.Open(fileLocation)
	if err != nil {
		return nil, fileSize, err
	}
	if strings.HasSuffix(fileLocation, ".mp4") || strings.HasSuffix(fileLocation, ".webm") || strings.HasSuffix(fileLocation, ".ogg") || strings.HasSuffix(fileLocation, ".mp3") {

		// Get the range header from the request
		rangeHeaders := c.GetReqHeaders()["range"]
		rangeHeader := ""
		if len(rangeHeaders) > 0 {
			rangeHeader = rangeHeaders[0]
		}
		if _debug {
			fmt.Printf("Range header: %s\n", rangeHeader)
		}

		if rangeHeader != "" {
			defer f.Close()
			// Get the range values
			rangeValues := strings.Split(rangeHeader, "=")[1]
			if _debug {
				fmt.Printf("Range values: %s\n", rangeValues)
			}
			// Get the start and end values
			byteRanges := strings.Split(rangeValues, "-")

			// get the start range
			start, err := strconv.ParseInt(byteRanges[0], 10, 64)
			if err != nil {
				log.Println("Error parsing start byte position:", err)
				return nil, 0, c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}
			var end int64
			// Calculate the end range
			if len(byteRanges) > 1 && byteRanges[1] != "" {
				end, err = strconv.ParseInt(byteRanges[1], 10, 64)
				if err != nil {
					log.Println("Error parsing end byte position:", err)
					return nil, 0, c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
				}
			} else {
				// maximum 1MB
				end = int64(math.Min(float64(start+1048576), float64(fileSize-1)))
			}

			// Setting required response headers
			c.Set(fiber.HeaderContentRange, fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize)) // Set the Content-Range header
			c.Set(fiber.HeaderContentLength, strconv.FormatInt(end-start+1, 10))                 // Set the Content-Length header for the range being served
			c.Set(fiber.HeaderAcceptRanges, "bytes")                                             // Set Accept-Ranges
			ctx.StatusCode = fiber.StatusPartialContent                                          // Set the status code to 206 (Partial Content)

			// Read the file from the start to the end
			_, err = f.Seek(start, 0)
			if err != nil {
				log.Println("Error seeking file:", err)
				return nil, 0, c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}
			_, err = io.CopyN(c, f, end-start+1)
			if err != nil {
				log.Println("Error copying file:", err)
				return nil, 0, c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}
			return nil, fileSize, nil
		} else {
			// Set the Content-Length header
			c.Set(fiber.HeaderContentLength, strconv.FormatInt(fileSize, 10))
			// Set the Accept-Ranges header
			c.Set(fiber.HeaderAcceptRanges, "bytes")
			// _, copyErr := io.Copy(c.Response().BodyWriter(), f)
			// if copyErr != nil {
			// 	log.Println("Error copying entire file to response:", copyErr)
			// 	return nil, 0, c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			// }
			// return nil, fileSize, nil
			// return nil, fileSize, c.SendStream(f)
			return io.Reader(f), fileSize, nil
		}

	} else {
		// io.Reader
		if fileSize >= 1048576 {
			return io.Reader(f), fileSize, nil
		} else {
			defer f.Close()
			fileContent = make([]byte, fileSize)
			_, err = f.Read(fileContent)
		}
	}
	return fileContent, fileSize, err
}

func SendStaticAsset(ctx *model.EventContext, c *fiber.Ctx) error {
	basePath := ctx.Data.GetString("base_path")
	path := ctx.Data.GetString("path")
	if _debug {
		fmt.Printf("Path: %s\n", path)
		fmt.Printf("Base path: %s\n", basePath)
	}

	fileLocation := convertHttpPathToFileLocation(basePath, path)

	if _debug {
		fmt.Printf("File location: %s\n", fileLocation)
	}
	var fileContent interface{}
	var err error
	var fileSize int64
	fileContent, fileSize, err = readFileBytes(ctx, c, fileLocation)
	ctx.Ephemeral = &model.EphemeralData{"fileSize": fileSize}
	if err != nil {
		return err
	}
	if ctx.StatusCode == 0 {
		ctx.StatusCode = 200
	}
	ctx.Result = fileContent
	return err
}

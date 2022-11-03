package sarteam

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	statusCheckNumRequests = 50
	statusCheckStartTile   = 150
)

func reportNetworkStatus(c *gin.Context, value bool) {
	c.JSON(http.StatusOK, gin.H{
		"online": value,
	})
}

func downloadTest(url string, ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	readBytes := 0
	buf := make([]byte, 1024)
	for {
		n, err := res.Body.Read(buf)
		readBytes += n
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
	}

	if readBytes < 128 {
		return fmt.Errorf("downloaded %d bytes, expected at least 128", readBytes)
	}

	return nil
}

func NetworkStatus(c *gin.Context) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	serverReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://sartopo.com/map.html", nil)
	if err != nil {
		reportNetworkStatus(c, false)
		return
	}

	serverResp, err := http.DefaultClient.Do(serverReq)
	if err != nil {
		reportNetworkStatus(c, false)
		return
	}

	defer serverResp.Body.Close()

	errorCount := 0
	for i := 0; i < statusCheckNumRequests; i++ {
		tileURL := fmt.Sprintf("https://sartopo.com/tile/bdem/10/160/%d.png", statusCheckStartTile+i)
		err = downloadTest(tileURL, ctx)
		if err != nil {
			errorCount++
		}
	}

	reportNetworkStatus(c, errorCount < statusCheckNumRequests/2) // Less than half the requests failed
}

func init() {
	APIRouter.GET("/api/networkStatus", NetworkStatus)
}

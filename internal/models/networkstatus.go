package models

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	statusCheckNumRequests = 50
	statusCheckStartTile   = 150
)

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

// InternetAvailable returns true if the internet is available.
func InternetAvailable() bool {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	serverReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://sartopo.com/map.html", nil)
	if err != nil {
		return false
	}

	serverResp, err := http.DefaultClient.Do(serverReq)
	if err != nil {
		return false
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

	return errorCount < statusCheckNumRequests/2 // Less than half the requests failed
}

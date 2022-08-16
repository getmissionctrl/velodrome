package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type cloudflareIps struct {
	IPV4 []string
	IPV6 []string
}

func GetCloudflareIPs(ctx context.Context) (*cloudflareIps, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.cloudflare.com/ips-v6", nil)
	if err != nil {
		return nil, err
	}
	ipv6Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		e := ipv6Resp.Body.Close()
		fmt.Println(e)
	}()
	req, err = http.NewRequestWithContext(ctx, "GET", "https://www.cloudflare.com/ips-v4", nil)
	if err != nil {
		return nil, err
	}
	ipv4Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		e := ipv4Resp.Body.Close()
		fmt.Println(e)
	}()

	if ipv4Resp.StatusCode != 200 || ipv6Resp.StatusCode != 200 {

		return nil, fmt.Errorf("cloudflare IP endpoints could not be reached: %s, %s", ipv4Resp.Status, ipv6Resp.Status)
	}

	b, err := io.ReadAll(ipv6Resp.Body)
	if err != nil {
		return nil, err
	}
	b2, err := io.ReadAll(ipv4Resp.Body)
	if err != nil {
		return nil, err
	}

	ips := cloudflareIps{}
	ipv6str := string(b)
	ipv6Lines := strings.Split(ipv6str, "\n")
	ips.IPV6 = append(ips.IPV6, ipv6Lines...)

	ipv4str := string(b2)
	fmt.Println(ipv4str)
	ipv4Lines := strings.Split(ipv4str, "\n")
	ips.IPV4 = append(ips.IPV4, ipv4Lines...)

	return &ips, nil
}

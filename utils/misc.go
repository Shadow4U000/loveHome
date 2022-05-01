package utils

func AddDomain2Url(url string) (domain_url string) {
	domain_url = "http://" + G_fdfs_http_addr + "/" + url
	return domain_url
}

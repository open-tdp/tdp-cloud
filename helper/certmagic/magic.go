package certmagic

import (
	"context"
	"errors"
	"strings"

	"github.com/caddyserver/certmagic"

	"tdp-cloud/cmd/args"
	"tdp-cloud/helper/logman"
	"tdp-cloud/helper/strutil"
)

var (
	sMagic = map[string]*certmagic.Config{}
	dMagic = map[string]*certmagic.Config{}
)

func Manage(rq *Params) error {

	skey := strutil.Md5(rq.Email + rq.SecretKey + rq.CaType)

	magic, ok := sMagic[skey]

	if !ok {
		magic = CreateMagic()
		magic.Issuers = []certmagic.Issuer{
			certmagic.NewACMEIssuer(magic, *newIssuer(rq)),
		}
		// 写入缓存
		sMagic[skey] = magic
	}

	dMagic[rq.Domain] = magic

	domains := strings.Split(rq.Domain, ",")
	return magic.ManageAsync(context.Background(), domains)

}

func Unmanage(domain string) {

	magic, ok := dMagic[domain]
	domains := strings.Split(domain, ",")

	if ok {
		magic.Unmanage(domains)
		delete(dMagic, domain)
	}

}

func CreateMagic() *certmagic.Config {

	config := certmagic.Config{
		Storage: &certmagic.FileStorage{
			Path: args.Dataset.Dir + "/certmagic",
		},
		Logger:  logman.Named("cert.magic"),
		OnEvent: magicEvent,
	}

	cache := certmagic.NewCache(certmagic.CacheOptions{
		GetConfigForCert: func(cert certmagic.Certificate) (*certmagic.Config, error) {
			return &config, nil
		},
		Logger: logman.Named("cert.cache"),
	})

	return certmagic.New(cache, config)

}

func CertDetail(domain string) (*Certificate, error) {

	cert := &Certificate{}

	magic, ok := dMagic[domain]

	if !ok {
		return cert, errors.New("域名不存在")
	}

	crt, err := magic.CacheManagedCertificate(context.Background(), domain)

	if err != nil {
		return cert, err
	}

	pk, err := certmagic.PEMEncodePrivateKey(crt.Certificate.PrivateKey)

	cert.Names = crt.Names
	cert.NotAfter = crt.Leaf.NotAfter.Unix()
	cert.NotBefore = crt.Leaf.NotBefore.Unix()
	cert.Certificate = crt.Certificate.Certificate
	cert.PrivateKey = pk

	cert.Issuer = map[string]any{
		"CommonName":   crt.Leaf.Issuer.CommonName,
		"Organization": crt.Leaf.Issuer.Organization[0],
		"Country":      crt.Leaf.Issuer.Country[0],
	}

	return cert, err

}

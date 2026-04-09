package service

import (
	"auth-haven/internal/domain/interfaces"
	"fmt"
	"net/mail"
	"strings"
)

type domainChecker struct {
	publicDomains map[string]bool
}

func NewDomainChecker() interfaces.DomainChecker {
	return &domainChecker{
		publicDomains: getPublicDomainList(),
	}
}

func (d *domainChecker) IsPublicDomain(email string) bool {
	address, err := mail.ParseAddress(email)
	if err != nil {
		return true // Treat invalid emails as public to be safe
	}

	parts := strings.Split(address.Address, "@")
	if len(parts) != 2 {
		return true
	}

	domain := strings.ToLower(parts[1])
	
	// Check exact match
	if d.publicDomains[domain] {
		return true
	}
	
	// Check subdomains of public domains
	for publicDomain := range d.publicDomains {
		if strings.HasSuffix(domain, "."+publicDomain) {
			return true
		}
	}
	
	return false
}

func (d *domainChecker) ExtractDomain(email string) (string, error) {
	address, err := mail.ParseAddress(email)
	if err != nil {
		return "", fmt.Errorf("invalid email format: %w", err)
	}

	parts := strings.Split(address.Address, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid email format")
	}

	return strings.ToLower(parts[1]), nil
}

func getPublicDomainList() map[string]bool {
	// List of common public email providers
	domains := []string{
		"gmail.com",
		"yahoo.com",
		"hotmail.com",
		"outlook.com",
		"icloud.com",
		"aol.com",
		"protonmail.com",
		"tutanota.com",
		"mail.com",
		"gmx.com",
		"yandex.com",
		"qq.com",
		"163.com",
		"126.com",
		"sina.com",
		"sohu.com",
		"zoho.com",
		"inbox.com",
		"mail.ru",
		"rambler.ru",
		"list.ru",
		"bk.ru",
		"ya.ru",
		"live.com",
		"msn.com",
		"passport.com",
		"comcast.net",
		"verizon.net",
		"att.net",
		"sbcglobal.net",
		"bellsouth.net",
		"charter.net",
		"cox.net",
		"earthlink.net",
		"juno.com",
		"netzero.com",
		"frontiernet.net",
		"windstream.net",
		"centurylink.net",
		"wowway.com",
		"twc.com",
		"optonline.net",
		"optum.com",
		"frontier.com",
		"suddenlink.net",
		"wowway.com",
		"rcn.com",
		"mediacomcc.com",
		"myfairpoint.net",
		"mybrighthouse.com",
		"wavecable.com",
		"astound.net",
		"grande.net",
		"hawaiiantel.net",
		"breezeline.com",
	}

	domainMap := make(map[string]bool)
	for _, domain := range domains {
		domainMap[domain] = true
	}
	
	return domainMap
}


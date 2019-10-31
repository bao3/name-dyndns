// Package dyndns provides a tool for running a
// dynamic dns updating service.
package dyndns

import (
	"fmt"
	"github.com/razoralpha/name-dyndns/api"
	"github.com/razoralpha/name-dyndns/log"
	"sync"
	"time"
)

var wg sync.WaitGroup

func contains(c api.Config, val string) bool {
	for _, v := range c.Hostnames {
		// We have a special case where an empty hostname
		// is equivalent to the domain (i.e. val == domain).
		if val == c.Domain && v == "" {
			return true
		} else if fmt.Sprintf("%s.%s.", v, c.Domain) == val {
			return true
		}
	}
	return false
}

func runConfig(c api.Config, daemon bool) {
	defer wg.Done()

	a := api.NewAPIFromConfig(c)
	for {
		ip, err := GetExternalIP()
		ipv6, errv6 := GetExternalIPv6()
		if err != nil {
			log.Logger.Print("Failed to retreive IPv4: ")
			log.Logger.Print(err)
			if daemon {
				log.Logger.Printf("Will retry in %d seconds...\n", c.Interval)
				time.Sleep(time.Duration(c.Interval) * time.Second)
				continue
			} else {
				log.Logger.Println("Giving up.")
				break
			}
		}
		log.Logger.Print("Retrieved IPv4: %s", ip)
		if errv6 != nil {
			log.Logger.Print("Failed to retreive IPv6: ")
			log.Logger.Print(err)
			if daemon {
				log.Logger.Printf("Will retry in %d seconds...\n", c.Interval)
				time.Sleep(time.Duration(c.Interval) * time.Second)
				continue
			} else {
				log.Logger.Println("Giving up.")
				break
			}
		}
		log.Logger.Print("Retrieved IPv6: %s", ipv6)

		// GetRecords retrieves a list of DNSRecords,
		// 1 per hostname with the associated domain.
		// If the content is not the current IP, then
		// update it.
		records, err := a.GetDNSRecords(c.Domain)
		if err != nil {
			log.Logger.Printf("Failed to retreive records for %s:\n\t%s\n", c.Domain, err)
			if daemon {
				log.Logger.Printf("Will retry in %d seconds...\n", c.Interval)
				time.Sleep(time.Duration(c.Interval) * time.Second)
				continue
			} else {
				log.Logger.Print("Giving up.")
				break
			}
		}

		for _, r := range records {
			log.Logger.Printf("Checking against %s (%s - %s)", r.FQDN, r.Type, r.Answer)
			if !contains(c, r.FQDN) {
				continue
			}

			if r.Type == "A" && r.Answer != ip {
				r.Answer = ip
				log.Logger.Printf("Updating %s with %s (ipv4)", r.Host, r.Answer)
				err = a.UpdateDNSRecord(r)
			} else if r.Type == "AAAA" && r.Answer != ipv6 {
				r.Answer = ipv6
				log.Logger.Printf("Updating %s with %s (ipv6)", r.Host, r.Answer)
				err = a.UpdateDNSRecord(r)
			}

			if err != nil {
				log.Logger.Printf("Failed to update record [%s] with IP: %s\n\t%s\n", r.FQDN, r.Answer, err)
			} else {
				log.Logger.Printf("Updated record [%s] with IP: %s\n", r.FQDN, r.Answer)
			}
		}

		log.Logger.Println("Update complete.")
		if !daemon {
			log.Logger.Println("Non daemon mode, stopping.")
			return
		}
		log.Logger.Printf("Will update again in %d seconds.\n", c.Interval)

		time.Sleep(time.Duration(c.Interval) * time.Second)
	}
}

// Run will process each configuration in configs.
// If daemon is true, then Run will run forever,
// processing each configuration at its specified
// interval.
//
// Each configuration represents a domain with
// multiple hostnames.
func Run(configs []api.Config, daemon bool) {
	for _, config := range configs {
		wg.Add(1)
		go runConfig(config, daemon)
	}

	wg.Wait()
}

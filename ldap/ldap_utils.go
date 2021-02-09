/*
Copyright 2021 Digitalis.IO.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ldap

import (
	"crypto/tls"
	"fmt"
	"os"

	ldapv1 "ldap-accounts-controller/api/v1"

	ldap "github.com/go-ldap/ldap/v3"
)

var (
	ldapFilter string = getEnv("LDAP_FILTER", "(objectClass=*)")
	baseDN     string = getEnv("LDAP_BASE_DN", "dc=axonops,dc=com")
)

func getEnv(key string, def string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		fmt.Printf("%s not set\n", key)
		return def
	} else {
		fmt.Printf("%s=%s\n", key, val)
	}

	return val
}

func LdapConnect() (*ldap.Conn, error) {
	ldapHostname := getEnv("LDAP_HOSTNAME", "localhost")
	ldapPort := getEnv("LDAP_PORT", "389")
	ldapBind := getEnv("LDAP_BIND", "cn=admin")
	ldapPassword := getEnv("LDAP_PASSWORD", "letmein")
	ldapTls := getEnv("LDAP_TLS", "false")

	var conn *ldap.Conn
	var err error
	if ldapTls == "true" {
		tlsConfig := tls.Config{
			InsecureSkipVerify: true,
			// ServerName: "the-target-server-of-ad",
			// RootCAs:    rootCA,
		}

		conn, err = ldap.DialTLS("tcp", fmt.Sprintf("%s:%s", ldapHostname, ldapPort), &tlsConfig)
	} else {
		conn, err = ldap.Dial("tcp", fmt.Sprintf("%s:%s", ldapHostname, ldapPort))
	}
	conn.Debug = true
	if err != nil {
		return nil, err
	}
	if err := conn.Bind(ldapBind, ldapPassword); err != nil {
		return nil, fmt.Errorf("Failed to bind. %s", err)
	}
	//defer conn.Close()

	return conn, nil
}

// LdapGet find a user or group from ldap server
func LdapGet(key string, value string) (ldapv1.LdapUserSpec, error) {
	conn, err := LdapConnect()
	if err != nil {
		return ldapv1.LdapUserSpec{}, fmt.Errorf("Could not connect to ldap server %s", err)
	}
	result, err := conn.Search(ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("%s=%s", key, value),
		[]string{},
		nil,
	))

	if err != nil {
		return ldapv1.LdapUserSpec{}, fmt.Errorf("Failed to search users. %s", err)
	}
	for _, entry := range result.Entries {
		fmt.Println(entry)
	}
	// not found
	if len(result.Entries) < 1 {
		return ldapv1.LdapUserSpec{}, nil
	}

	if key == "username" {
		var userAccount = ldapv1.LdapUserSpec{
			Username: result.Entries[0].GetAttributeValue("uid"),
			GID:      result.Entries[0].GetAttributeValue("uidNumber"),
			UID:      result.Entries[0].GetAttributeValue("gidNumber"),
			Shell:    result.Entries[0].GetAttributeValue("shell"),
			Homedir:  result.Entries[0].GetAttributeValue("homeDirectory"),
		}
		return userAccount, nil
	}
	return ldapv1.LdapUserSpec{}, nil
}

func LdapDeleteUser(user ldapv1.LdapUserSpec) error {
	if user.Username == "" {
		fmt.Println("User not found")
		return nil
	}
	conn, err := LdapConnect()
	if err != nil {
		return fmt.Errorf("Could not connect to ldap server %s", err)
	}

	dn := fmt.Sprintf("uid=%s,ou=People,%s", user.Username, baseDN)
	delReq := ldap.NewDelRequest(dn, []ldap.Control{})
	fmt.Printf("%v\n", delReq)
	if err := conn.Del(delReq); err != nil {
		return err
	}

	return nil
}

func LdapAddUser(user ldapv1.LdapUserSpec) error {
	conn, err := LdapConnect()
	if err != nil {
		return fmt.Errorf("Could not connect to ldap server %s", err)
	}

	dn := fmt.Sprintf("uid=%s,ou=People,%s", user.Username, baseDN)
	addReq := ldap.NewAddRequest(dn, []ldap.Control{})

	addReq.Attribute("objectClass",
		[]string{"top", "posixAccount", "shadowAccount", "account"})

	addReq.Attribute("uid", []string{user.Username})
	addReq.Attribute("cn", []string{user.Username})
	addReq.Attribute("uidNumber", []string{user.UID})
	addReq.Attribute("gidNumber", []string{user.GID})
	addReq.Attribute("homeDirectory", []string{user.Homedir})
	addReq.Attribute("gecos", []string{user.Username})
	addReq.Attribute("userPassword", []string{user.Password})
	addReq.Attribute("loginShell", []string{user.Shell})

	fmt.Printf("%v\n", addReq)
	if err := conn.Add(addReq); err != nil {
		return err
	}

	return nil
}

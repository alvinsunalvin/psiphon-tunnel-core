// +build android

/*
 * Copyright (c) 2015, Psiphon Inc.
 * All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package psiphon

import (
	"errors"
	"net"
	"strings"

	"github.com/Psiphon-Inc/openssl"
)

// newOpenSSLConn wraps a connection with TLS which mimicks stock Android TLS.
// This facility is used as a circumvention measure to ensure Psiphon client
// TLS ClientHello messages match common TLS ClientHellos vs. the more
// distinguishable (blockable) Go TLS ClientHello.
func newOpenSSLConn(rawConn net.Conn, config *CustomTLSConfig) (handshakeConn, error) {

	if !config.SkipVerify {
		return nil, ContextError(errors.New("opensslDial certificate verification not supported"))
	}
	if config.SendServerName {
		return nil, ContextError(errors.New("opensslDial server name not supported"))
	}

	ctx, err := openssl.NewCtx()
	if err != nil {
		return nil, ContextError(err)
	}

	// Use the same cipher suites, in the same priority order, as stock Android TLS.
	// Based on: https://android.googlesource.com/platform/external/conscrypt/+/master/src/main/java/org/conscrypt/NativeCrypto.java
	// This list includes include recently retired DSS suites: https://android.googlesource.com/platform/external/conscrypt/+/e53baea9221be7f9828d0f338ede284e22f55722%5E!/#F0,
	// as those are still commonly deployed.
	ciphersuites := []string{
		"ECDHE-ECDSA-AES128-GCM-SHA256",
		"ECDHE-ECDSA-AES256-GCM-SHA384",
		"ECDHE-RSA-AES128-GCM-SHA256",
		"ECDHE-RSA-AES256-GCM-SHA384",
		"DHE-RSA-AES128-GCM-SHA256",
		"DHE-RSA-AES256-GCM-SHA384",
		"ECDHE-ECDSA-AES128-SHA",
		"ECDHE-ECDSA-AES256-SHA",
		"ECDHE-RSA-AES128-SHA",
		"ECDHE-RSA-AES256-SHA",
		"DHE-RSA-AES128-SHA",
		"DHE-RSA-AES256-SHA",
		"DHE-DSS-AES128-SHA",
		"DHE-DSS-AES256-SHA",
		"ECDHE-ECDSA-RC4-SHA",
		"ECDHE-RSA-RC4-SHA",
		"AES128-GCM-SHA256",
		"AES256-GCM-SHA384",
		"AES128-SHA",
		"AES256-SHA",
		"RC4-SHA",
	}
	ctx.SetCipherList(strings.Join(ciphersuites, ":"))

	// Mimic extensions used by stock Android.
	// NOTE: Heartbeat extension is disabled at compile time.
	ctx.SetOptions(openssl.NoSessionResumptionOrRenegotiation | openssl.NoTicket)

	conn, err := openssl.Client(rawConn, ctx)
	if err != nil {
		return nil, ContextError(err)
	}

	return conn, nil
}

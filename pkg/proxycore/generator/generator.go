package generator

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/rsa"
    "crypto/tls"
    "crypto/x509"
    "crypto/x509/pkix"
    "encoding/pem"
    "io/ioutil"
    "log"
    "math/big"
    "time"
    "github.com/jctanner/pkgproxy/pkg/proxycore/utils"
)

// Load your CA certificate and key
var (
    caCert = loadCACert(utils.GetEnvOrDefault("CACERT", "/src/caCert.pem"))
    caKey  = loadCAKey(utils.GetEnvOrDefault("CAKEY", "/src/caKey.pem"))
)

// GetCertificateFunc returns a function compatible with tls.Config's GetCertificate.
// It dynamically generates a certificate for the given ClientHelloInfo, signed by your CA.
func GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
    return func(helloInfo *tls.ClientHelloInfo) (*tls.Certificate, error) {
        // Generate a private key for the new certificate
        priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
        if err != nil {
            return nil, err
        }

        // Create a certificate template
        template := x509.Certificate{
            SerialNumber: big.NewInt(1), // SerialNumber should be unique
            Subject: pkix.Name{
                Organization: []string{"Your MITM Proxy"},
            },
            NotBefore:             time.Now(),
            NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year validity
            KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
            ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
            BasicConstraintsValid: true,
        }

        // Generate the certificate, signed by your CA
        certDER, err := x509.CreateCertificate(rand.Reader, &template, caCert, &priv.PublicKey, caKey)
        if err != nil {
            return nil, err
        }

        // Create a tls.Certificate with the generated certificate and private key
        cert := tls.Certificate{
            Certificate: [][]byte{certDER},
            PrivateKey:  priv,
        }

        return &cert, nil
    }
}

// loadCACert and loadCAKey are placeholders for functions that would load your CA's certificate and key.
// You'll need to implement these based on how you store your CA certificate and key (e.g., PEM files).

// loadCACert loads a CA certificate from a specified file path.
func loadCACert(certPath string) *x509.Certificate {
    certPEM, err := ioutil.ReadFile(certPath)
    if err != nil {
        log.Fatalf("Failed to read CA certificate file: %v", err)
    }
    block, _ := pem.Decode(certPEM)
    if block == nil {
        log.Fatalf("Failed to decode CA certificate PEM")
    }
    cert, err := x509.ParseCertificate(block.Bytes)
    if err != nil {
        log.Fatalf("Failed to parse CA certificate: %v", err)
    }
    return cert
}

func loadCAKey(keyPath string) interface{} {
    keyPEM, err := ioutil.ReadFile(keyPath)
    if err != nil {
        log.Fatalf("Failed to read private key file: %v", err)
    }

    block, _ := pem.Decode(keyPEM)
    if block == nil {
        log.Fatalf("Failed to decode PEM block containing private key")
    }

    key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
    if err != nil {
        log.Fatalf("Failed to parse private key: %v", err)
    }

    switch key := key.(type) {
    case *rsa.PrivateKey:
        log.Println("Loaded an RSA private key.")
        return key
    case *ecdsa.PrivateKey:
        log.Println("Loaded an ECDSA private key.")
        return key
    default:
        log.Fatalf("Unsupported key type found.")
    }
    // This return won't be reached but is necessary to compile.
    return nil
}

func GenerateTLSCertificateForHost(host string) (tls.Certificate, error) {
    // Generate a new private key for the certificate
    privKey, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        return tls.Certificate{}, err
    }

    // Set up a serial number for the certificate
    serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
    if err != nil {
        return tls.Certificate{}, err
    }

    // Create a certificate template
    certTemplate := x509.Certificate{
        SerialNumber: serialNumber,
        Subject: pkix.Name{
            Organization: []string{"Dynamic Certificates, Inc."},
            CommonName:   host,
        },
        NotBefore:             time.Now(),
        NotAfter:              time.Now().Add(24 * time.Hour), // Valid for 1 day
        KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
        ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
        BasicConstraintsValid: true,
    }

    // Create the certificate, signed with the CA's private key
    derBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, caCert, &privKey.PublicKey, caKey)
    if err != nil {
        return tls.Certificate{}, err
    }

    // Create a TLS certificate using the certificate and private key
    tlsCert := tls.Certificate{
        Certificate: [][]byte{derBytes},
        PrivateKey:  privKey,
    }

    return tlsCert, nil
}

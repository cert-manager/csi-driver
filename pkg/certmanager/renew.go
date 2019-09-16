package certmanager

type renewer struct {
	dataDir string
	nodeID  string
}

// do discovery (cert-manager-csi-$pod_name-$vol_id
// attempt to read certs and get expiry
// build go routines that sleep for expiry - RenewBefore
// when renew - find existing CertificateRequest
//  if exist get all options and delete old
//  if not exist, guess options based on the local cert
// create new CR with options
// overwrite new cert when created
// loop go routine to wait for next renewal

package gce_auto_discovery

// Ready implements the ready.Readiness interface, once this flips to true CoreDNS
// assumes this plugin is ready for queries; it is not checked again.
func (e GceAutoDiscovery) Ready() bool { return e.Ready() }

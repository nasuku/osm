package witesand

import (
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"

	"github.com/openservicemesh/osm/pkg/announcements"
	"github.com/openservicemesh/osm/pkg/kubernetes/events"
	"github.com/openservicemesh/osm/pkg/service"
)

func NewWitesandCatalog(kubeClient kubernetes.Interface, clusterId string) *WitesandCatalog {
	wc := WitesandCatalog{
		myIP:               "",
		masterOsmIP:        "",
		clusterId:          clusterId,
		remoteK8s:          make(map[string]RemoteK8s),
		clusterPodMap:      make(map[string]ClusterPods),
		allPodMap:          make(map[string]ClusterPods),
		kubeClient:         kubeClient,
		apigroupToPodMap:   make(map[string]ApigroupToPodMap),
		apigroupToPodIPMap: make(map[string]ApigroupToPodIPMap),
	}

	wc.UpdateMasterOsmIP()

	return &wc
}

// cache myIP
func (wc *WitesandCatalog) RegisterMyIP(ip string) {
	log.Info().Msgf("[RegisterMyIP] myIP:%s", ip)
	wc.myIP = ip
}

func (wc *WitesandCatalog) GetMyIP() string {
	return wc.myIP
}

// read env to update masterOsmIP (in case master OSM restarted)
func (wc *WitesandCatalog) UpdateMasterOsmIP() {
	// TODO revisit: may not need complicated mechanism, env do not change
	newIP := os.Getenv("MASTER_OSM_IP")
	if newIP != wc.masterOsmIP {
		log.Info().Msgf("[RegisterMasterOsmIP] masterOsmIP:%s", newIP)
		wc.UpdateRemoteK8s("master", newIP)
		wc.masterOsmIP = newIP
	}
}

func (wc *WitesandCatalog) IsMaster() bool {
	return wc.masterOsmIP == ""
}

// update the context with received remoteK8s
func (wc *WitesandCatalog) UpdateRemoteK8s(remoteClusterId string, remoteIP string) {
	if remoteClusterId == "" {
		return
	}

	// handle the case of remoteIP not responding, remove it from the list after certain retries
	if remoteIP == "" {
		remoteK8, exists := wc.remoteK8s[remoteClusterId]
		if exists {
			remoteK8.failCount += 1
			if remoteK8.failCount >= 3 {
				log.Info().Msgf("[UpdateRemoteK8s] Delete clusterId:%s", remoteClusterId)
				delete(wc.remoteK8s, remoteClusterId)
				wc.UpdateClusterPods(remoteClusterId, nil)
				wc.UpdateAllPods(remoteClusterId, nil)
				return
			}
			wc.remoteK8s[remoteClusterId] = remoteK8
		}
		return
	}

	log.Info().Msgf("[UpdateRemoteK8s] IP:%s clusterId:%s", remoteIP, remoteClusterId)
	remoteK8, exists := wc.remoteK8s[remoteClusterId]
	if exists {
		if remoteK8.OsmIP != remoteIP {
			log.Info().Msgf("[UpdateRemoteK8s] update IP:%s clusterId:%s", remoteIP, remoteClusterId)
			// IP address changed ?
			wc.remoteK8s[remoteClusterId] = RemoteK8s{
				OsmIP: remoteIP,
			}
		}
	} else {
		log.Info().Msgf("[UpdateRemoteK8s] create IP:%s clusterId:%s", remoteIP, remoteClusterId)
		wc.remoteK8s[remoteClusterId] = RemoteK8s{
			OsmIP: remoteIP,
		}
	}
}

func (wc *WitesandCatalog) ListRemoteK8s() map[string]RemoteK8s {
	// TODO LOCK
	remoteK8s := wc.remoteK8s

	return remoteK8s
}

func (wc *WitesandCatalog) IsWSGatewayService(svc service.MeshServicePort) bool {
	return strings.HasPrefix(svc.Name, "gateway")
}

func (wc *WitesandCatalog) updateEnvoy() {
	events.GetPubSubInstance().Publish(events.PubSubMessage{
		AnnouncementType: announcements.ScheduleProxyBroadcast,
		NewObj:           nil,
		OldObj:           nil,
	})
}

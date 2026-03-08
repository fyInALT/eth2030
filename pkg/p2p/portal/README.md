# Package portal

Portal Network implementation: content-addressed DHT for history, state, and beacon data with LRU content storage, XOR distance routing, eclipse-attack resistance, and content-radius-based placement.

## Overview

The `portal` package implements the Ethereum Portal Network protocol for lightweight data access. `ContentDB` provides an in-memory LRU store with XOR-distance-based content radius management: the node's radius shrinks as storage fills, reducing its share of the global content space. `DHTRouter` wraps a `discover.KademliaTable` with Portal-specific content routing — iterative `RouteContentRequest` queries converge on the node that holds a given content ID, and `OfferContent` propagates new content to nearby peers.

`HistoryNetwork` serves EIP-4444 block/receipt data; `StateNetwork` serves state trie data. `EclipseResistance` implements Kademlia eclipse-attack mitigation. `ContentValidator` verifies content proofs before storage.

## Functionality

### ContentDB

- `NewContentDB(config ContentDBConfig) *ContentDB`
- `Get(id ContentID) ([]byte, error)` / `Put(id ContentID, data []byte) error` / `Delete(id ContentID) error` / `Has(id ContentID) bool`
- `StoreContentByKey(contentKey, content []byte) error` — checks radius before storing
- `FindContentByKey(contentKey []byte) ([]byte, error)`
- `EntriesWithinRadius(nodeID NodeID, radius NodeRadius) []ContentID`
- `FarthestContent() (ContentID, *big.Int)`
- `SetRadius(r NodeRadius)` / `Radius() NodeRadius` / `AutoUpdateRadius()`
- `UsedBytes()` / `CapacityBytes()` / `ItemCount()`

### DHTRouter

- `NewDHTRouter(table *discover.KademliaTable, config DHTRouterConfig) *DHTRouter`
- `RouteContentRequest(contentID [32]byte, queryFn DHTQueryFunc) (*ContentResponse, error)`
- `IterativeNodeLookup(target [32]byte, queryFn ...) []discover.NodeEntry`
- `FindContentProviders(contentID [32]byte, maxProviders int) []discover.NodeEntry`
- `OfferContent(contentKeys [][32]byte, offerFn ...) (int, error)`
- `UpdateRadius(totalStored, maxStorage uint64)`
- `IsWithinRadius(contentID [32]byte) bool`
- `ComputeContentDist(nodeID, contentID [32]byte) *big.Int`

### Helpers

- `ContentKeyToID(contentKey []byte) ContentID` — keccak256
- `IsWithinRadius(nodeID NodeID, contentID ContentID, radius NodeRadius) bool`
- `GossipContent(contentKey, content []byte, peers []*PeerInfo, nodeID NodeID) (*GossipResult, error)`
- `DistanceMetric(nodeID NodeID, contentID ContentID) *big.Int`
- `UpdateRadiusFromUsage(usedBytes, capacityBytes uint64) NodeRadius`

## Usage

```go
db := portal.NewContentDB(portal.DefaultContentDBConfig(localNodeID))
db.StoreContentByKey(contentKey, blockData)

router := portal.NewDHTRouter(table, portal.DefaultDHTRouterConfig())
resp, err := router.RouteContentRequest(contentID, func(node discover.NodeEntry, id [32]byte) ([]byte, []byte, []discover.NodeEntry, error) {
    return queryPeer(node, id)
})
```

[← p2p](../README.md)

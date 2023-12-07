# apicache

API Cache is a Go package that gives a cluster of servers hosting an API the ability to cache API responses. Its primary use cases are:
* De-duplicating requests that must be idempotent
* Caching responses for expensive operations so they don't have to be recomputed.

API Cache relies on Redis as the underlying infrastructure.

#### Install
```
go get github.com/danielwchapman/apicache
```

### Main Concepts

#### Idempotency
Say you want to ensure an API operation is idempotent, which means, that if you do the operation multiple times, it has the same effect as doing it once. Many operations, like simply updating a value, are intrinsically idempotent. Other operations, like creating a resource or incrementing a counter, often need a developer to implement idempotency explicitly.

One way to achieve idempotency is to ask the client to send a unique Request ID, like a UUID, with each request. That way, when the server receives the same RequestID multiple times, it knows it should de-duplicate the request. Ideally, the API should return the same response for each duplicated call, since the reason the client is sending multiple calls is probably because it did not get the first response.

#### Response Cache
Caching a response is straightforward conceptually, but potentially complex to implement perfectly. The reason is if the underlying data store is changed, the cache *might* have to be invalidated. Since the cache key is generally a hash of a combination of API parameters, it's usually not easy to invalidate the cache entry immediately when it becomes stale. For this reason, this style of cache should only be used when the response doesn't need to contain the most up-to-date data.

One sweet spot for this style of cache is to populate dashboards where the data is updated fairly regularly, but it's acceptable if the dashboard is not updated in real time. You may choose to update the dashboard every, say, 5 minutes instead of in real time.

#### Usage

Towards the beginning of your API handler, call `ReceiveAndWait`. The outcomes that can happen are:
* If the RequestID parameter has been seen before, it returns the response for the previously computed request with a matching RequestID.
* If the RquestID has not been seen within a given time window, it returns a `FirstSeen` status. It's then up to the API handler to call `Handle` and provide the response to cache.
* If the RequestID has been seen before, but hasn't been handled, this function will block. It then either:
  * Return the response once it has been received by another request operation.
  * Return context deadline exceeded error (a time out)
* There is simply an error - ex: Redis server could not be reached.
    

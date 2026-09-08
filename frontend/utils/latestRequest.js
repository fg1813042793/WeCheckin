export function createLatestRequestTracker() {
  let latestID = 0

  return {
    begin() {
      latestID += 1
      return latestID
    },
    isLatest(requestID) {
      return requestID === latestID
    },
    invalidate() {
      latestID += 1
    }
  }
}

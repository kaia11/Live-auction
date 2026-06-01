package realtime

// AtomicBidScript is the stage-C draft for the future Redis Lua bid path.
// The current project still runs the bid decision in memory, but this script
// fixes the contract we will implement when the Redis client is wired in.
const AtomicBidScript = `
-- KEYS[1] session state hash
-- KEYS[2] session ranking zset
-- KEYS[3] session participants set
-- KEYS[4] bid request dedup key
-- ARGV[1] request id
-- ARGV[2] user id
-- ARGV[3] bid price
-- ARGV[4] now unix seconds
-- ARGV[5] dedup ttl seconds

local cached = redis.call("GET", KEYS[4])
if cached then
  local cachedObj = cjson.decode(cached)
  cachedObj.code = "duplicate_request"
  return cjson.encode(cachedObj)
end

local status = redis.call("HGET", KEYS[1], "status")
if status ~= "bidding" then
  return cjson.encode({
    ok = false,
    code = "session_not_bidding"
  })
end

local currentPrice = tonumber(redis.call("HGET", KEYS[1], "current_price") or "0")
local incrementStep = tonumber(redis.call("HGET", KEYS[1], "increment_step") or "0")
local startPrice = tonumber(redis.call("HGET", KEYS[1], "start_price") or "0")
local extensionSeconds = tonumber(redis.call("HGET", KEYS[1], "extension_seconds") or "0")
local extensionTriggerSeconds = tonumber(redis.call("HGET", KEYS[1], "extension_trigger_seconds") or "0")
local endTimeUnix = tonumber(redis.call("HGET", KEYS[1], "end_time_unix") or "0")
local bidPrice = tonumber(ARGV[3])
local nowUnix = tonumber(ARGV[4])
local nextMinimum = currentPrice + incrementStep

if currentPrice == 0 and bidPrice < startPrice then
  return cjson.encode({
    ok = false,
    code = "invalid_bid_price"
  })
end

if bidPrice < nextMinimum then
  return cjson.encode({
    ok = false,
    code = "invalid_bid_price"
  })
end

if incrementStep > 0 and ((bidPrice - currentPrice) % incrementStep ~= 0) then
  return cjson.encode({
    ok = false,
    code = "invalid_bid_price"
  })
end

local ceilingPriceRaw = redis.call("HGET", KEYS[1], "ceiling_price")
local acceptedBidPrice = bidPrice
local ceilingReached = false
if ceilingPriceRaw and ceilingPriceRaw ~= "" then
  local ceilingPrice = tonumber(ceilingPriceRaw)
  if bidPrice >= ceilingPrice then
    acceptedBidPrice = ceilingPrice
    ceilingReached = true
  end
end

local extensionApplied = false
if not ceilingReached and endTimeUnix > 0 then
  if (endTimeUnix - nowUnix) <= extensionTriggerSeconds then
    endTimeUnix = endTimeUnix + extensionSeconds
    extensionApplied = true
  end
end

redis.call("HSET", KEYS[1],
  "current_price", acceptedBidPrice,
  "leader_user_id", ARGV[2],
  "end_time_unix", endTimeUnix
)
redis.call("SADD", KEYS[3], ARGV[2])
redis.call("ZADD", KEYS[2], acceptedBidPrice, ARGV[2])

local participantCount = redis.call("SCARD", KEYS[3])
redis.call("HSET", KEYS[1], "participant_count", participantCount)

local rank = redis.call("ZREVRANK", KEYS[2], ARGV[2])
local response = {
  ok = true,
  code = "",
  accepted_bid_price = acceptedBidPrice,
  current_price = acceptedBidPrice,
  participant_count = participantCount,
  rank = rank and (rank + 1) or 0,
  ceiling_reached = ceilingReached,
  extension_applied = extensionApplied,
  end_time_unix = endTimeUnix,
  next_minimum_bid = acceptedBidPrice + incrementStep
}

redis.call("SETEX", KEYS[4], tonumber(ARGV[5]), cjson.encode(response))

return cjson.encode(response)
`

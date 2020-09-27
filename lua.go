package ranking

const (
	luaTemplate = `
local ZKEY, HKEY = KEYS[1]..":z", KEYS[1]..":h"
local f = function() %s end
local o = cmsgpack.unpack(table.remove(ARGV, 1))
if o then
	if o.construct_from and redis.call("EXISTS", ZKEY) == 0 and
		redis.call("EXISTS", o.construct_from..":z") == 1 then
		redis.call("ZUNIONSTORE", ZKEY, 1, o.construct_from..":z")
		local v = redis.call("HGETALL", o.construct_from..":h")
		if #v > 0 then redis.call("HMSET", HKEY, unpack(v)) end
	end
	local r = f()
	if o.capacity then
		local size = redis.call("ZCARD", ZKEY)
		if size > o.capacity then
			local v = redis.call("ZPOPMIN", ZKEY, size - o.capacity)
			local a = {}
			for i=1,#v-1,2 do a[#a+1] = v[i] end
			redis.call("HDEL", HKEY, unpack(a))
		end
	end
	if o.expire_at then
		redis.call("PEXPIREAT", ZKEY, o.expire_at)
		redis.call("PEXPIREAT", HKEY, o.expire_at)
	elseif o.idle_expire then
		redis.call("PEXPIRE", ZKEY, o.idle_expire)
		redis.call("PEXPIRE", HKEY, o.idle_expire)
	else
		redis.call("PERSIST", ZKEY)
		redis.call("PERSIST", HKEY)
	end
	return r
else
	return f()
end`
)

var (
	luaTouch = newScript(``)

	luaRemoveId = newScript(`
redis.call("HDEL", HKEY, unpack(ARGV))
return redis.call("ZREM", ZKEY, unpack(ARGV))`)

	/// leaderboard
	luaSetScore = newScript(`
local es = cmsgpack.unpack(ARGV[1])
if #es == 0 then return 0 end
local za = {}
local ha = {}
for _, e in ipairs(es) do
	za[#za+1], za[#za+2] = e.score, e.id
	if e.info and e.info ~= "" then
		ha[#ha+1], ha[#ha+2] = e.id, e.info
	end
end
if #ha > 0 then redis.call("HSET", HKEY, unpack(ha)) end
return redis.call("ZADD", ZKEY, unpack(za))`)

	luaIncScore = newScript(`
local es = cmsgpack.unpack(ARGV[1])
if #es == 0 then return 0 end
local a = {}
local r = {}
for i, e in ipairs(es) do
	r[i] = tonumber(redis.call("ZADD", ZKEY, "INCR", e.score, e.id))
	if e.info and e.info ~= "" then
		a[#a+1], a[#a+2] = e.id, e.info
	end
end
if #a > 0 then redis.call("HSET", HKEY, unpack(a)) end
return cmsgpack.pack(r)`)

	luaSetInfo = newScript(`
local es = cmsgpack.unpack(ARGV[1])
if #es == 0 then return 0 end
local a = {}
for _, e in ipairs(es) do a[#a+1], a[#a+2] = e.id, e.info end
return redis.call("HSET", HKEY, unpack(a))`)

	luaGetRange = newScript(`
local es = {}
local r = redis.call("ZREVRANGE", ZKEY, ARGV[1], ARGV[2], "WITHSCORES")
if #r == 0 then return cmsgpack.pack(es) end
local a = {}
for i=1,#r-1,2 do
	es[#es + 1] = { ["id"] = r[i], ["score"] = tonumber(r[i+1]) }
	a[#a + 1] = r[i]
end
local r = redis.call("HMGET", HKEY, unpack(a))
for i=1,#r,1 do
	es[i].rank = ARGV[1] + i - 1
	if r[i] then es[i].info = r[i] end
end
return cmsgpack.pack(es)`)

	luaGetById = newScript(`
local e = { ["id"] = ARGV[1] }
e.rank = redis.call("ZREVRANK", ZKEY, ARGV[1])
if not e.rank then return nil end
e.score = assert(tonumber(redis.call("ZSCORE", ZKEY, ARGV[1])))
e.info = redis.call("HGET", HKEY, ARGV[1])
if not e.info then e.info = nil end
return cmsgpack.pack(e)`)

	// bubble
	luaAppend = newScript(`
local es = cmsgpack.unpack(ARGV[1])
if #es == 0 then return 0 end
local n = 0
local r = redis.call("ZRANGE", ZKEY, 0, 0, "WITHSCORES")
if r and #r == 2 then n = r[2] end
local za = {}
local ha = {}
for _, e in ipairs(es) do
	if not redis.call("ZSCORE", ZKEY, e.id) then
		n = n - 1
		za[#za+1], za[#za+2] = n, e.id
		if e.info and e.info ~= "" then
			ha[#ha+1], ha[#ha+2] = e.id, e.info
		end
	end
end
if #za == 0 then return 0 end
if #ha > 0 then redis.call("HSET", HKEY, unpack(ha)) end
return redis.call("ZADD", ZKEY, unpack(za))`)

	luaSwapById = newScript(`
local s1 = redis.call("ZSCORE", ZKEY, ARGV[1])
local s2 = redis.call("ZSCORE", ZKEY, ARGV[2])
if not s1 and not s2 then return 0 end
if s1 and not s2 then return redis.call("ZADD", ZKEY, s1, ARGV[2]) end
if not s1 and s2 then return redis.call("ZADD", ZKEY, s2, ARGV[1]) end
return redis.call("ZADD", ZKEY, s2, ARGV[1], s1, ARGV[2])`)

	luaSwapByRank = newScript(`
local r1 = redis.call("ZREVRANGE", ZKEY, ARGV[1], ARGV[1], "WITHSCORES")
if not s1 then error('rank "' .. ARGV[1] .. '" not found') end
local r2 = redis.call("ZREVRANGE", ZKEY, ARGV[2], ARGV[2], "WITHSCORES")
if not s1 then error('rank "' .. ARGV[2] .. '" not found') end
return redis.call("ZADD", ZKEY, r2[2], r1[1], r1[2], r2[1])`)
)

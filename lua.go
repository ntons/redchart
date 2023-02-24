package redchart

import "github.com/ntons/redis"

func newScript(src string) *redis.Script {
	return redis.NewScript(luaTemplate, src)
}

const (
	luaTemplate = `
local ZKEY, HKEY = KEYS[1]..":z", KEYS[1]..":h"
local o = cmsgpack.unpack(table.remove(ARGV, 1))
local f = function() %s end
if o then
	if o.construct_from and redis.call("EXISTS", ZKEY) == 0 and redis.call("EXISTS", o.construct_from..":z") == 1 then
		redis.call("ZUNIONSTORE", ZKEY, 1, o.construct_from..":z")
		if not o.no_info then
			local v = redis.call("HGETALL", o.construct_from..":h")
			if #v > 0 then redis.call("HMSET", HKEY, unpack(v)) end
		end
	end
	local r = f()
	if o.capacity and o.capacity > 0 and not o.no_trim then
		local size = redis.call("ZCARD", ZKEY)
		if size > o.capacity then
			local v = redis.call("ZPOPMIN", ZKEY, size - o.capacity)
			local a = {}
			for i=1,#v-1,2 do a[#a+1] = v[i] end
			redis.call("HDEL", HKEY, unpack(a))
		end
	end
	if o.persist then
		redis.call("PERSIST", ZKEY)
		redis.call("PERSIST", HKEY)
	elseif o.expire_at then
		redis.call("EXPIREAT", ZKEY, o.expire_at)
		redis.call("EXPIREAT", HKEY, o.expire_at)
	elseif o.idle_expire then
		redis.call("EXPIRE", ZKEY, o.idle_expire)
		redis.call("EXPIRE", HKEY, o.idle_expire)
	end
	return r
else
	return f()
end`
)

// assume that, N elements in chart, M elements to work out
var (
	luaTouch = newScript(``)

	// O(M*log(N))
	luaRemoveId = newScript(`
redis.call("HDEL", HKEY, unpack(ARGV))
return redis.call("ZREM", ZKEY, unpack(ARGV))`)

	/// leaderboard

	// O(M*log(N))
	luaSet = newScript(`
local elist = cmsgpack.unpack(ARGV[1])
if elist == nil or #elist == 0 then return cmsgpack.pack({}) end

local zargs = {}
if o and o.set then
	if o.set.only_add then zargs[#zargs+1] = "NX" end
	if o.set.only_update then zargs[#zargs+1] = "XX" end
	if o.set.incr_by then zargs[#zargs+1] = "INCR" end
end
local zoptn = #zargs

-- 更新榜单
local function update(e)
	if o and o.capacity and o.capacity > 0 and o.no_trim and redis.call("ZCARD", ZKEY) >= o.capacity and not redis.call("ZSCORE", ZKEY, e.id) then e.rank, e.score = -1, 0 return end
	zargs[zoptn+1], zargs[zoptn+2], e.rank = e.score, e.id, -2
	redis.call("ZADD", ZKEY, unpack(zargs))
end
for _, e in ipairs(elist) do update(e) end

-- 获取更新后榜单
local function retrieve(e)
	if e.rank == -1 then return end
	local score = redis.call("ZSCORE", ZKEY, e.id)
	if not score then e.rank, e.score = -1, 0 return end
	local rank = redis.call("ZREVRANK", ZKEY, e.id)
	if o and o.capacity and o.capacity > 0 and rank >= o.capacity then e.rank, e.score = -1, 0 return end
	e.score, e.rank = tonumber(score), rank
end
for _, e in ipairs(elist) do retrieve(e) end

-- 设置详细信息
local function update_info(e)
	if e.rank < 0 then e.info = nil return end
	if e.info and e.info ~= "" then
		redis.call("HSET", HKEY, e.id, e.info)
		if o.no_info then e.info = nil end
	elseif not o.no_info then
		local info = redis.call("HGET", HKEY, e.id)
		if info then e.info = info end
	end
end
for _, e in ipairs(elist) do update_info(e) end

-- 返回更新后数据
return cmsgpack.pack(elist)`)

	// O(M*log(N))
	luaIncr = newScript(`
local es = cmsgpack.unpack(ARGV[1])
if #es == 0 then return 0 end
local a = {}
local r = {}
for i, e in ipairs(es) do
	r[i] = tonumber(redis.call("ZADD", ZKEY, "INCR", e.score, e.id))
	if e.info and e.info ~= "" then a[#a+1], a[#a+2] = e.id, e.info end
end
if #a > 0 then redis.call("HSET", HKEY, unpack(a)) end
return cmsgpack.pack(r)`)

	//
	luaRandByScore = newScript(`
local r = {}
for i, a in ipairs(cmsgpack.unpack(ARGV[1])) do
    local min_id = redis.call("ZRANGEBYSCORE", ZKEY, a.min, a.max, "LIMIT", 0, 1)
	local max_id = redis.call("ZREVRANGEBYSCORE", ZKEY, a.max, a.min, "LIMIT", 0, 1)
	if #min_id == 1 and #max_id == 1 then
		local min_rk = redis.call("ZRANK", ZKEY, min_id[1])
		local max_rk = redis.call("ZRANK", ZKEY, max_id[1])
		if max_rk - min_rk + 1 <= a.count then
			local x = redis.call("ZRANGE", ZKEY, min_rk, max_rk)
			for i=1, #x do r[#r+1] = x[i] end
		else
		    math.randomseed(redis.call("TIME")[2])
			local u = {}
			for i=1,a.count do
			    local rk = 0
			    repeat rk = math.random(min_rk, max_rk) until not u[rk]
				u[rk] = 1
				r[#r+1] = redis.call("ZRANGE", ZKEY, rk, rk)[1]
			end
		end
	end
end
return cmsgpack.pack(r)`)

	// O(M)
	luaSetInfo = newScript(`
local es = cmsgpack.unpack(ARGV[1])
if #es == 0 then return 0 end
local a = {}
for _, e in ipairs(es) do a[#a+1], a[#a+2] = e.id, e.info end
return redis.call("HSET", HKEY, unpack(a))`)

	// O(log(N)+M)
	luaGetByRank = newScript(`
local r = {}
local x = redis.call("ZREVRANGE", ZKEY, ARGV[1], ARGV[2], "WITHSCORES")
local a = {}
for i=1,#x-1,2 do
	r[#r+1] = { ["rank"] = ARGV[1] + #r, ["id"] = x[i], ["score"] = tonumber(x[i+1]) }
	a[#a+1] = x[i]
end
if #a > 0 and not o.no_info then
	local x = redis.call("HMGET", HKEY, unpack(a))
	for i=1,#x,1 do if x[i] then r[i].info = x[i] end end
end
return cmsgpack.pack(r)`)

	// O(M*log(N))
	// first ZSCORE[O(1)], then ZREVRANK[O(logN)]
	luaGetById = newScript(`
local r = {}
for _, id in ipairs(ARGV) do
    local x = redis.call("ZSCORE", ZKEY, id)
	if x then
		local e = { ["id"] = id, ["score"] = tonumber(x), ["rank"] = redis.call("ZREVRANK", ZKEY, id) }
	    if not o.no_info then
			local info = redis.call("HGET", HKEY, id)
			if info then e.info = info end
		end
		r[#r+1] = e
	end
end
return cmsgpack.pack(r)`)

	// bubble
	// O(M*log(N))
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
		if e.info and e.info ~= "" then ha[#ha+1], ha[#ha+2] = e.id, e.info end
	end
end
if #za == 0 then return 0 end
if #ha > 0 then redis.call("HSET", HKEY, unpack(ha)) end
return redis.call("ZADD", ZKEY, unpack(za))`)

	// O(log(N))
	luaSwapById = newScript(`
local s1 = redis.call("ZSCORE", ZKEY, ARGV[1])
local s2 = redis.call("ZSCORE", ZKEY, ARGV[2])
if not s1 and not s2 then return 0 end
if s1 and not s2 then return redis.call("ZADD", ZKEY, s1, ARGV[2]) end
if not s1 and s2 then return redis.call("ZADD", ZKEY, s2, ARGV[1]) end
return redis.call("ZADD", ZKEY, s2, ARGV[1], s1, ARGV[2])`)

	// O(log(N))
	luaSwapByRank = newScript(`
local r1 = redis.call("ZREVRANGE", ZKEY, ARGV[1], ARGV[1], "WITHSCORES")
if not r1 then error('rank "' .. ARGV[1] .. '" not found') end
local r2 = redis.call("ZREVRANGE", ZKEY, ARGV[2], ARGV[2], "WITHSCORES")
if not r1 then error('rank "' .. ARGV[2] .. '" not found') end
return redis.call("ZADD", ZKEY, r2[2], r1[1], r1[2], r2[1])`)
)

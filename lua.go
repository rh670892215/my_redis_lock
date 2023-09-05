package my_redis_lock

const LuaCheckAndDelDistributionLock = `
	local lockerKey = KEYS[1]
	local lockerToken = ARGV[1]
	local getToken = redis.call('get',lockerKey)
	if(not getToken or getToken ~= lockerToken)then
		return 0
	else
		return redis.call('del',lockerKey)
	end
`

const LuaDelayDistributionLockExpire = `
local lockerKey = KEYS[1]
local lockerToken = ARGV[1]
local duration = ARGV[2]
local getToken = redis.call('get',lockerKey)
if(not getToken or getToken ~= lockerToken)then
	return 0
else
	return redis.call('expire',lockerKey,duration)
end
`

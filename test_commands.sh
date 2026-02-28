#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "🚀 Starting GoRedis Command Battery Test..."

# PING
RES=$(redis-cli -p 6369 PING)
if [ "$RES" == "PONG" ]; then echo -e "${GREEN}PASS: PING${NC}"; else echo -e "${RED}FAIL: PING ($RES)${NC}"; fi

# SET/GET
redis-cli -p 6369 SET project "GoRedis" > /dev/null
RES=$(redis-cli -p 6369 GET project)
if [ "$RES" == "GoRedis" ]; then echo -e "${GREEN}PASS: SET/GET${NC}"; else echo -e "${RED}FAIL: SET/GET ($RES)${NC}"; fi

# INCR - reset state first
redis-cli -p 6369 DEL visitor_count > /dev/null
RES=$(redis-cli -p 6369 INCR visitor_count)
if [ "$RES" == "1" ]; then echo -e "${GREEN}PASS: INCR (new)${NC}"; else echo -e "${RED}FAIL: INCR new ($RES)${NC}"; fi

RES=$(redis-cli -p 6369 INCR visitor_count)
if [ "$RES" == "2" ]; then echo -e "${GREEN}PASS: INCR (existing)${NC}"; else echo -e "${RED}FAIL: INCR existing ($RES)${NC}"; fi

# DEL
redis-cli -p 6369 SET to_delete "gone" > /dev/null
RES=$(redis-cli -p 6369 DEL to_delete)
if [ "$RES" == "1" ]; then echo -e "${GREEN}PASS: DEL${NC}"; else echo -e "${RED}FAIL: DEL ($RES)${NC}"; fi

# TTL
redis-cli -p 6369 SET temp_key "value" EX 3 > /dev/null
sleep 1
RES=$(redis-cli -p 6369 TTL temp_key)
if [ "$RES" == "2" ]; then echo -e "${GREEN}PASS: TTL${NC}"; else echo -e "${RED}FAIL: TTL ($RES)${NC}"; fi

# Expiration
sleep 3
RES=$(redis-cli -p 6369 GET temp_key)
if [ -z "$RES" ]; then echo -e "${GREEN}PASS: Expiration${NC}"; else echo -e "${RED}FAIL: Expiration (Key still exists)${NC}"; fi

echo "🏁 Test Suite Finished!"
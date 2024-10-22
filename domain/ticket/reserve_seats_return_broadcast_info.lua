-- Get input arguments
local eventID = ARGV[1]
local sectionID = ARGV[2]
local rowID = ARGV[3]
local length = tonumber(ARGV[4])
local sessionID = ARGV[5] -- for tracking user session

-- Key for seat availability
local seatsKey = "event:" .. eventID .. ":section:" .. sectionID .. ":rows"

-- Get the row data
local rowData = redis.call("HGET", seatsKey, rowID)
if not rowData then
    return {err = "Row not found"}
end

local rowInfo = cjson.decode(rowData)
local seats = rowInfo.seats

local reservedSeats = {} -- Empty table, can be used as array, dict, ...
local seatCount = #seats -- # Get the length of string
local consecutiveCount = 0
local maxConsecutive = 0

-- Mark the reserved seats 
for i = 1, seatCount do
    local seat = seats:sub(i, i)
    if seat == "0" then -- Available 
        consecutiveCount = consecutiveCount + 1
    else
        consecutiveCount = 0 -- Reset 
    end

    -- Reserve the seat if enough consecutive seats are found
    -- Marks backwards
    if consecutiveCount == length then
        for j = 1, length do
            reservedSeats[i - j + 1] = true -- Mark reserved seats
        end
        break
    end
end

-- Check if we found enough consecutive seats
if next(reservedSeats) == nil then -- next() used to iterate table, it table is empty it returns nil
    return {err = "Not enough consecutive seats available"}
end

-- Reserve the seats
for index, _ in pairs(reservedSeats) do -- pairs() used to iterate over all key-value pairs in the table
    seats = seats:sub(1, index - 1) .. "1" .. seats:sub(index + 1) -- Mark seat as reserved
end

-- Update the row data
rowInfo.seats = seats
redis.call("HSET", seatsKey, rowID, cjson.encode(rowInfo))

-- ======================================================== 

-- Calculate max consecutive lengths for each price block
local priceBlocksKey = "event:" .. eventID .. ":section:" .. sectionID .. ":price_blocks"
local priceMaxConsecutive = {}

local priceBlocks = redis.call("ZRANGE", priceBlocksKey, 0, -1, "WITHSCORES")
-- return odd: member, even:score

-- Filter price blocks by rowID
local filteredPriceBlocks = {}

for i = 1, #priceBlocks, 2 do -- 2 -> i += 2
    local block = priceBlocks[i] -- member
    local price = tonumber(priceBlocks[i + 1]) -- score

     -- Check if the block's row_id matches the given rowID
    local currentRowID = block:match("^(%d+):") -- Extract row_id from block
    if currentRowID == rowID then
        table.insert(filteredPriceBlocks, block)
        table.insert(filteredPriceBlocks, price) -- Add the corresponding price
    end
end

-- Initialize priceMaxConsecutive with zero lengths for filtered blocks
for i = 1, #filteredPriceBlocks, 2 do
    local block = filteredPriceBlocks[i]
    local price = tonumber(filteredPriceBlocks[i + 1])
    priceMaxConsecutive[block] = 0 -- Start with 0 for each price block
end

-- Check availability in the row for each filtered price block
for i = 1, #filteredPriceBlocks, 2 do
    local block = filteredPriceBlocks[i]
    local price = tonumber(filteredPriceBlocks[i + 1])

    -- Get start and end seat IDs and numbers from the block
    local start_seat_id, start_seat_number, end_seat_id, end_seat_number = block:match(":(%d+):(%d+):(%d+):(%d+)")
    start_seat_number = tonumber(start_seat_number)
    end_seat_number = tonumber(end_seat_number)

    -- Check availability within the defined range of seats
    local maxLength = 0
    local currentLength = 0

    for j = start_seat_number, end_seat_number do
        local seat = seats:sub(j, j)
        if seat == "0" then -- Available seat
            currentLength = currentLength + 1
        else
            if currentLength > maxLength then
                maxLength = currentLength
            end
            currentLength = 0
        end
    end

    -- Final check for the last segment
    if currentLength > maxLength then
        maxLength = currentLength
    end

    -- Update max length for the current price block
    priceMaxConsecutive[block] = maxLength
end

-- Filter to only keep the maximum lengths for each price
local finalMaxConsecutive = {}
for price, length in pairs(priceMaxConsecutive) do
    finalMaxConsecutive[price] = length
end

return {reserved = reservedSeats, maxConsecutive = finalMaxConsecutive}

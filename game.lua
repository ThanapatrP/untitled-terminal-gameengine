-- game.lua : platformer แบบ side-scrolling สำหรับ Fluid (ASCII ขาวดำ)
--
-- ปุ่ม:  ซ้าย/ขวา = เดิน   z หรือ ขึ้น = กระโดด (กดค้างกระโดดสูง ปล่อยเร็วกระโดดต่ำ)
--        x = เล่นใหม่หลังจบด่าน   q = ออก
--
-- แก้ด่านได้ที่ LEVEL ข้างล่าง ทุกตัวอักษรคือหนึ่งช่อง:
--   #  บล็อกตัน          =  แพลตฟอร์ม (ตันเหมือนกัน)
--   *  เหรียญ            ^  หนาม (โดนแล้วตาย)
--   E  ศัตรู (เหยียบหัวได้)  S  จุดเริ่ม (ตำแหน่งเท้า)
--   |  >  ~  ตัวอื่นๆ เป็นของตกแต่ง เดินทะลุได้
-- ถึงคอลัมน์ GOAL_X (เสาธง) = ชนะ

local LEVEL = {
  "",
  "",
  "                                                             ~~~~",
  "        ~~~~                                                                                                             ~~~~",
  "                                                                                          ~~~~",
  "                                  ~~~~                                                                                                            ~~~~",
  "",
  "",
  "",
  "",
  "",
  "",
  "                                                                                         * * *",
  "                                                                                        =======",
  "",
  "                                                                                                                                                      |>>",
  "                                * * *                                                                E                  *                             |",
  "                     ***      =========                              *        =======             =======            =======                          |",
  "                                                           #        ===                                                           **                  |",
  "                                                          ##                                                                                          |",
  "   S    * * *                               ^^^     E    ###                                                  E     ^^^^^^^^^               E         |",
  "####################     #########################################       #######################################################      ##########################",
  "####################     #########################################       #######################################################      ##########################",
}

local GOAL_X      = 150
local SCREEN_W    = 64
local MAP_TOP     = 1      -- แถว 0 ของจอเป็น HUD, แผนที่เริ่มแถว 1

-- ฟิสิกส์ (หน่วย: ช่องต่อเฟรม, 30 fps)
-- ช่องตัวอักษรสูงราว 2 เท่าของความกว้าง ค่าแนวตั้งจึงเล็กกว่าแนวนอน
local RUN         = 0.45
local GRAVITY     = 0.06
local JUMP_V      = -0.75  -- กระโดดได้สูงราว 4 ช่อง
local MAX_FALL    = 0.9    -- ต้องน้อยกว่า 1 ไม่งั้นตกทะลุพื้นได้
local ENEMY_SPEED = 0.12

local map, map_w, map_h
local start_x, start_y
local p, enemies
local coins, total_coins, deaths
local state, dead_timer, frame

-- ---------- map ----------

local function tile(x, y)
  if x < 0 or x >= map_w then return "#" end   -- ขอบซ้าย/ขวาเป็นกำแพง
  if y < 0 or y >= map_h then return " " end
  return map[y + 1]:sub(x + 1, x + 1)
end

local function set_tile(x, y, ch)
  local row = map[y + 1]
  map[y + 1] = row:sub(1, x) .. ch .. row:sub(x + 2)
end

local function solid(x, y)
  local t = tile(x, y)
  return t == "#" or t == "="
end

local function spike(x, y) return tile(x, y) == "^" end

-- ตรวจทุกช่องที่กล่อง (x, y, w, h) ทับอยู่ คืนช่องแรกที่ test เป็นจริง
local function box_hits(x, y, w, h, test)
  for ty = math.floor(y), math.floor(y + h - 0.001) do
    for tx = math.floor(x), math.floor(x + w - 0.001) do
      if test(tx, ty) then return true, tx, ty end
    end
  end
  return false
end

local function load_level()
  map, enemies = {}, {}
  map_h, map_w = #LEVEL, SCREEN_W
  total_coins = 0
  for _, row in ipairs(LEVEL) do
    if #row > map_w then map_w = #row end
  end
  for y, row in ipairs(LEVEL) do
    row = row .. string.rep(" ", map_w - #row)
    local out = {}
    for x = 1, map_w do
      local c = row:sub(x, x)
      if c == "S" then
        start_x, start_y = x - 1, y - 2         -- S คือเท้า ตัวละครสูง 2 ช่อง
        c = " "
      elseif c == "E" then
        enemies[#enemies + 1] = { x = x - 1, y = y - 1, dir = -1, alive = true }
        c = " "
      elseif c == "*" then
        total_coins = total_coins + 1
      end
      out[x] = c
    end
    map[y] = table.concat(out)
  end
end

-- ---------- game state ----------

local function respawn()
  p = { x = start_x, y = start_y, vy = 0, ground = false, moving = false }
end

local function new_game()
  load_level()
  coins, deaths, frame = 0, 0, 0
  state = "play"
  respawn()
end

local function kill_player()
  deaths = deaths + 1
  state = "dead"
  dead_timer = 30
end

-- ---------- update ----------

local function update_player()
  -- เดินซ้ายขวา
  local vx = 0
  if btn(0) then vx = -RUN end
  if btn(1) then vx = RUN end
  p.moving = vx ~= 0

  local nx = p.x + vx
  if box_hits(nx, p.y, 1, 2, solid) then
    if vx > 0 then nx = math.floor(nx) else nx = math.floor(nx) + 1 end
  end
  p.x = nx

  -- กระโดด
  local jump = btn(4) or btn(2)
  if jump and p.ground then p.vy = JUMP_V end
  if not jump and p.vy < -0.2 then p.vy = -0.2 end  -- ปล่อยปุ่มเร็ว = กระโดดต่ำ

  p.vy = math.min(p.vy + GRAVITY, MAX_FALL)
  local ny = p.y + p.vy
  p.ground = false
  if p.vy > 0 then
    local hit, _, ty = box_hits(p.x, ny, 1, 2, solid)
    if hit then ny = ty - 2; p.vy = 0; p.ground = true end
  elseif p.vy < 0 then
    local hit, _, ty = box_hits(p.x, ny, 1, 2, solid)
    if hit then ny = ty + 1; p.vy = 0 end
  end
  p.y = ny

  -- ตกเหว / โดนหนาม
  if p.y > map_h or box_hits(p.x, p.y, 1, 2, spike) then
    kill_player()
    return
  end

  -- เก็บเหรียญ
  box_hits(p.x, p.y, 1, 2, function(tx, ty)
    if tile(tx, ty) == "*" then
      set_tile(tx, ty, " ")
      coins = coins + 1
    end
    return false
  end)

  if p.x >= GOAL_X then state = "win" end
end

local function update_enemies()
  for _, e in ipairs(enemies) do
    if e.alive then
      -- เดินกลับตัวเมื่อเจอกำแพง หนาม หรือขอบ
      local nx = e.x + e.dir * ENEMY_SPEED
      local front = (e.dir > 0) and math.floor(nx + 0.999) or math.floor(nx)
      if solid(front, e.y) or spike(front, e.y) or not solid(front, e.y + 1) then
        e.dir = -e.dir
      else
        e.x = nx
      end

      -- ชนผู้เล่น
      local top, bottom = math.floor(p.y), math.floor(p.y + 1.999)
      if math.abs(e.x - p.x) < 0.8 and e.y >= top and e.y <= bottom then
        if p.vy > 0 and (p.y + 2) - e.y < 1.0 then
          e.alive = false                 -- เหยียบหัว
          p.vy = -0.5
        else
          kill_player()
          return
        end
      end
    end
  end
end

function _init()
  new_game()
end

function _update()
  frame = frame + 1
  if state == "play" then
    update_player()
    if state == "play" then update_enemies() end
  elseif state == "dead" then
    dead_timer = dead_timer - 1
    if dead_timer <= 0 then
      respawn()
      state = "play"
    end
  elseif state == "win" then
    if btn(5) then new_game() end
  end
end

-- ---------- draw ----------

function _draw()
  cls()

  local px = math.floor(p.x + 0.5)
  local cam = math.max(0, math.min(map_w - SCREEN_W, px - 28))

  for y = 1, map_h do
    print(map[y]:sub(cam + 1, cam + SCREEN_W), 0, y - 1 + MAP_TOP)
  end

  local step = math.floor(frame / 6) % 2
  for _, e in ipairs(enemies) do
    if e.alive then
      put(math.floor(e.x + 0.5) - cam, e.y + MAP_TOP, step == 0 and "M" or "W")
    end
  end

  local sx, sy = px - cam, math.floor(p.y) + MAP_TOP
  if state == "dead" then
    put(sx, sy, "x"); put(sx, sy + 1, "A")
  else
    local body = "A"
    if not p.ground then body = "Y"
    elseif p.moving and step == 0 then body = "H" end
    put(sx, sy, "o"); put(sx, sy + 1, body)
  end

  print(string.format(" COINS %d/%d   DEATHS %d ", coins, total_coins, deaths), 1, 0)

  if state == "win" then
    fill(14, 8, 49, 13, " ")
    print("+----------------------------------+", 14, 8)
    print("|             YOU WIN!             |", 14, 9)
    print(string.format("|      coins %2d/%-2d   deaths %-3d    |", coins, total_coins, deaths), 14, 10)
    print("|       press X to play again      |", 14, 11)
    print("+----------------------------------+", 14, 12)
  end
end

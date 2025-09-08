IconPath = "q.png"
Line = "Q"
Station = "R16"
Direction = "downtown"

QMenu = hs.menubar.new()
-- set the menubar as non-icon
Icon = hs.image.imageFromPath(IconPath)
Icon = Icon:setSize({ w = 16, h = 16 })
QMenu:setIcon(Icon, false)
QMenu:stateImageSize({ w = 16, h = 16 })

function WhenIsTheQ()
	local timeToQ = "??:??"
	local output, status, _, _ = hs.execute(
		"whenistheq next_train --station " .. Station .. " --line " .. Line .. " --diff --direction " .. Direction,
		true
	) -- 'r' for read mode to capture output
	if status then
		timeToQ = string.gsub(output, "[\r\n]+$", "")
	end
	return timeToQ
end

hs.timer.doEvery(1, function()
	local title = WhenIsTheQ()
	QMenu:setTitle(title)
end)

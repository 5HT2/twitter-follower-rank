on run argv
    try
        set limit to item 1 of argv as number
        if limit is equal to 0 then error "Number can't be zero"
    on error errMsg number errNum
        display dialog "Error: " & errMsg & " (" & errNum & ")"
        return
    end try

    tell application "Google Chrome" to activate

    delay 1

    tell application "System Events"
        set activeApp to name of first application process whose frontmost is true
        if activeApp does not contain "Google Chrome" then
            display dialog "Error: Google Chrome is not active"
            return
        end if

        keystroke "n=0"
        delay 0.1
        key code 36
        delay 0.15

        repeat with i from 0 to limit
            keystroke "followers[n].content.content = followers[n].content.#g"
            delay 0.1
            key code 36
            delay 0.15
            keystroke "n++"
            delay 0.1
            key code 36
            delay 0.15
        end repeat
    end tell
end run

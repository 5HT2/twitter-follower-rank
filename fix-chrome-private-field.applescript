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
        repeat 77 times
            keystroke "followers[n].content.content = followers[n].content.#g"
            delay 0.1
            key code 36
            delay 0.1
            keystroke "n++"
            delay 0.1
            key code 36
            delay 0.1
        end repeat
    end tell

end run

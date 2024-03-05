on run

    delay 2

    tell application "Google Chrome" to activate

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

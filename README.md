# twitter-follower-rank

A simple Go program to print a Twitter User's list of followers, sorted by number of followers.

### This does **not** use or require API access.

![](./.github/IMAGES/preview.png)

## Usage

```bash
# Get repo && cd
git clone https://github.com/5HT2/twitter-follower-rank && cd twitter-follower-rank
```
```bash
# Build
make
# Once you have a data.json, run:
./twitter-follower-rank
# That's it!
```

<!-- GENERATED FROM MAKEFILE -->
```
Usage of ./twitter-follower-rank:
  -f string
        Data file to read (default "data.json" or matching "data-[0-9]{4}-[0-9]{2}-[0-9]{2}(-following)?.json")
  -following
        Invert mutuals detection mode to the following tab instead of the followers tab
  -ratio
        Only display followers with a following:follower ratio of >= -ratioBuf
  -ratioBuf float
        Buffer for ranked. e.g. If set to 0.9, it will display if (followers / following >= 0.9) (default 0.9)
```
<!-- GENERATED FROM MAKEFILE -->

### Auto select `data.json` file

<details><summary>An explanation of how automatically selecting a data file works:</summary>

The dates given here are **examples**, to show how the sort order works.
- The flag `-f` is always prioritized as first, and is only ignored if the provided file cannot be read / does not exist.
- The `data.json` value is always **last** as a final fallback.
- The rest of the files are found from the existing files in the current working directory.

By default, the data file will by selected in this order:
```
my-custom-file.json (optionally set by -f)
data-2024-04-20.json
data-2024-03-01.json
data-2023-11-30.json
data.json
```

If you have the `-following` flag set it would instead look like this:
```
my-custom-file.json (optionally set by -f)
data-2024-04-20-following.json
data-2024-03-01-following.json
data.json
```

</details>

## What is a `data.json`

The way that this works is by leveraging Twitter's own Followers tab, and simply grabbing an object of all the requests.

In the future, I'll probably add a way to also do this via Twitter's GDPR data request, assuming it includes enough info to not have to scrape.
#### Pros

- No $$$ API key
- It just works

#### Cons
- Come up with some way to scroll down. I did it by hand.
- After ~2k followers I got rate-limited with a 429 response. Go make some tea and come back.

## How to get your very own `data.json`

1. To do so, open the Chrome Dev Tools with <kbd>Ctrl</kbd> <kbd>Shift</kbd> <kbd>I</kbd>, go to the network tab.
2. In the network tab, search the text `Followers?`. This will filter it to only the requests we want.
3. Now, open the followers menu on Twitter, or visit https://[domain]/username/followers.
4. Scroll all the way to the bottom, just keep in mind that going too fast will rate-limit you. If you hit the limit, just wait 30min - 1h.
4. Once at the bottom, open a _new_ Dev Tools window for your existing Dev Tools window with <kbd>Ctrl</kbd> <kbd>Shift</kbd> <kbd>I</kbd> (yes, debug inception).
5. In the new Dev Tools window, open the console tab (taken from [StackOverflow](https://stackoverflow.com/a/57782978), works on Chrome 111 or newer) and run this:
```javascript
let followers = await (async () => {
  const getContent = r => r.url() && !r.url().startsWith('data:') && r.contentData();
  const nodes = UI.panels.network.networkLogView.dataGrid.rootNode().flatChildren();
  const requests = nodes.map(n => n.request());
  const contents = await Promise.all(requests.map(getContent));
  return contents.map((data, i) => {
    const r = requests[i];
    const url = r.url();
    const body = data?.content;
    const content = !data ? url :
        r.contentType().isTextType() ? data :
            typeof body !== 'string' ? body :
                `data:${r.mimeType}${data.encoded ? ';base64' : ''},${body}`;
    return { url, content };
  });
})();
```

Once you've scrolled all the way to the bottom, <kbd>Right Click</kbd> the output of the last console command, and do "Copy Object".

![](./.github/IMAGES/chrome.png)

#### Congrats! Now you can create a file called `data.json` inside this project, and paste the JSON object into it.
The program itself handles all of the parsing and such from that point.

#### NOTE: Twitter attempted to patch this.
As of 2024/03, it looks like Twitter switched to storing the follower info from `followers[n].content.content` to `followers[n].content.#g`.

This is annoying, as it means the Chrome Dev tools don't include the final `.content` when doing "Copy Object", *but* there is a workaround.
There's a `fix-chrome-private-field.applescript` script which you can run with `osascript fix-chrome-private-field.applescript [number of items]`.

The script essentially just types out `followers[n].content.content = followers[n].content.#g` incrementally and increases `n`. You can't do this in a for loop in Chrome for some reason, as it will give you a private field access error.

If you'd like a Linux version of the script, or know how to get around Chrome not allowing accessing private fields in a for loop, feel free to open an issue or message me (my contact info is listed on [my profile](https://github.com/5HT2)).

## Disclaimer(s)

1. I have no idea if Twitter will care if you scroll down to the bottom of the followers tab, I'd advise you to tread with caution though.
2. This code was incredibly hastily written, it's kind of awful, easy to improve but it works, I don't care enough to bother. It does what I need it to.
3. This might break if the response format changes. Unlikely, but feel free to make a Github issue about it.
4. This data appears to be kind of cached on the server, if someone changes their username the old one is still sent in a response (multiple weeks after the username change). There's no error checking for empty profiles either.
For the most part, it's functional, with the rare 1 or 2 duplicate accounts, the follower counts might also be a lil outdated, this is just Twitter's caching / optimization stuff on their end.
5. Idk this is really janky, if it breaks or there's something you want me to fix, feel free to ask lol

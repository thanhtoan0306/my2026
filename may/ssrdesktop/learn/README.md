# Learn: SSR Web vs Desktop App

Short guides that explain concepts used in [`may/ssrdesktop`](../README.md). Read in order, or jump to what you need.

| # | Topic | File |
|---|--------|------|
| 1 | What is SSR? | [01-what-is-ssr.md](01-what-is-ssr.md) |
| 2 | SSR on the web vs SSR in a desktop app | [02-ssr-web-vs-desktop.md](02-ssr-web-vs-desktop.md) |
| 3 | How *this* repo combines both | [03-how-ssrdesktop-works.md](03-how-ssrdesktop-works.md) |
| 4 | Cheat sheet (one page) | [04-cheat-sheet.md](04-cheat-sheet.md) |
| 5 | Share the app with a friend | [05-sharing-with-a-friend.md](05-sharing-with-a-friend.md) |

## One-sentence summary

**SSR web** = HTML is built on a server and sent to a **browser** over the network.  
**SSR desktop (this project)** = HTML is still built on a **server process**, but the “browser” is a **native window** on your machine—and the server usually only listens on `127.0.0.1`, not the public internet.

The rendering step (Go + `html/template`) can look almost the same. What changes is **who opens the UI**, **where the server runs**, and **how the app is distributed**.

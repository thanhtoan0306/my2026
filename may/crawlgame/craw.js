const fs = require("fs");
const path = require("path");
const axios = require("axios");
const cheerio = require("cheerio");

const URL = "https://www.rsg-games.com/en-US/games.html";

async function crawlGames() {
  const { data: html } = await axios.get(URL, {
    headers: {
      "User-Agent":
        "Mozilla/5.0",
    },
  });

  const $ = cheerio.load(html);

  const games = [];
  const seen = new Set();

  // inspect actual game links only
  $("a").each((_, el) => {
    const href = $(el).attr("href") || "";

    // only keep TryGame links
    if (!href.includes("TryGame")) return;

    const card = $(el);

    const raw = card.text().replace(/\s+/g, " ").trim();

    // title
    let title =
      card.find(".game-name").text().trim() ||
      raw.split("Max. Multiply")[0]?.trim();

    // cleanup duplicated titles
    title = title.replace(/\s+/g, " ").trim();

    const parts = title.split(" ");

    // remove repeated duplicate name
    const half = Math.floor(parts.length / 2);

    if (
      parts.slice(0, half).join(" ") ===
      parts.slice(half).join(" ")
    ) {
      title = parts.slice(0, half).join(" ");
    }

    if (!title || seen.has(title)) return;

    seen.add(title);

    // HOT / NEW
    let tag = null;

    if (raw.includes("HOT")) tag = "HOT";
    if (raw.includes("NEW")) tag = "NEW";

    // multiplier
    const multiplier =
      raw.match(/Multiply\s*([0-9,X]+)/i)?.[1] || null;

    // rules
    const rules =
      raw.match(/Rules\s*(.+?)(Introduction|Free Trial|$)/i)?.[1]?.trim() ||
      null;

    // image
    let image =
      card.find("img").attr("src") || null;

    if (image && !image.startsWith("http")) {
      image = "https://www.rsg-games.com" + image;
    }

    // full try url
    const tryGameUrl = href.startsWith("http")
      ? href
      : "https://www.rsg-games.com" + href;

    // extract gameid
    const gameid =
      href.match(/gameid=(\d+)/i)?.[1] || null;

    games.push({
      gameid,
      title,
      tag,
      multiplier,
      rules,
      image,
      tryGameUrl,
    });
  });

  const outDir = path.join(
    process.cwd(),
    "@may/crawlgame"
  );

  fs.mkdirSync(outDir, { recursive: true });

  fs.writeFileSync(
    path.join(outDir, "games.json"),
    JSON.stringify(games, null, 2)
  );

  console.log(`Saved ${games.length} games`);
}

crawlGames();
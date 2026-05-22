const fs = require('fs').promises;
const path = require('path');

async function addGameIds() {
    const inputPath = path.join(__dirname, 'games1.json');
    const outputPath = path.join(__dirname, 'games2.json');

    const raw = await fs.readFile(inputPath, 'utf-8');
    const data = JSON.parse(raw);

    const updatedGames = data.games.map(game => {
        const match = game.url && game.url.match(/games_detail_(\d+)\.html/);
        const id_game = match ? parseInt(match[1], 10) : null;
        const trialUrl = id_game
            ? `https://www.rsg-games.com/TryGame?lang=en-US&gameid=${id_game}`
            : null;

        return {
            ...game,
            id_game,
            trialUrl
        };
    });

    const output = {
        ...data,
        games: updatedGames
    };

    await fs.writeFile(outputPath, JSON.stringify(output, null, 2), 'utf-8');
    console.log(`Done! Saved ${updatedGames.length} games to: ${outputPath}`);
}

addGameIds().catch(console.error);

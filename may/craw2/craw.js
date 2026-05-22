const { chromium } = require('playwright');
const fs = require('fs').promises;
const path = require('path');

async function crawlGames() {
    const url = 'https://www.rsg-games.com/en-US/games.html';
    
    console.log('Launching browser...');
    const browser = await chromium.launch({ headless: true });
    const page = await browser.newPage();
    
    try {
        console.log(`Navigating to ${url}...`);
        // Navigate to the target and wait for network activity to go idle
        await page.goto(url, { waitUntil: 'networkidle' });
        
        console.log('Waiting for game containers to load...');
        // Standard slot/casino dynamic sites usually use generic list elements. 
        // Adjust selectors below if the class names differ upon site updates.
        await page.waitForSelector('.game-item, .game-list-item, a[href*="game"]', { timeout: 15000 })
            .catch(() => console.log("Timeout waiting for specific card selectors, attempting generic scrape."));

        // Evaluate code directly inside the browser context to parse elements safely
        const gamesList = await page.evaluate(() => {
            // Modify these generic selectors matching RSG's DOM grid system
            // Common targets: cards, list elements, or links containing game images
            const items = document.querySelectorAll('.game-item, .game-card, .list-item, .portfolio-item');
            const data = [];
            
            if (items.length > 0) {
                items.forEach(item => {
                    const titleEl = item.querySelector('.title, h3, h4, .game-name');
                    const imgEl = item.querySelector('img');
                    const linkEl = item.querySelector('a');
                    const categoryEl = item.querySelector('.category, .tag');

                    data.push({
                        title: titleEl ? titleEl.innerText.trim() : 'Unknown Game',
                        thumbnail: imgEl ? imgEl.src : null,
                        url: linkEl ? linkEl.href : null,
                        category: categoryEl ? categoryEl.innerText.trim() : 'General'
                    });
                });
            } else {
                // Fallback: If elements are wrapped inside structural links (A tags)
                const fallbackLinks = document.querySelectorAll('main a, #app a, .content a');
                fallbackLinks.forEach(link => {
                    const img = link.querySelector('img');
                    if (img && (link.href.includes('game') || img.alt)) {
                        data.push({
                            title: img.alt || 'Game Unit',
                            thumbnail: img.src,
                            url: link.href,
                            category: 'Slot/Game'
                        });
                    }
                });
            }
            
            return data;
        });

        console.log(`Successfully scraped ${gamesList.length} games.`);
        
        // Output formatting
        const outputPayload = {
            sourceUrl: url,
            scrapedAt: new Date().toISOString(),
            totalGames: gamesList.length,
            games: gamesList
        };

        const outputPath = path.join(__dirname, 'games.json');
        await fs.writeFile(outputPath, JSON.stringify(outputPayload, null, 2), 'utf-8');
        console.log(`JSON array created successfully at: ${outputPath}`);

    } catch (error) {
        console.error('An error occurred during execution:', error);
    } finally {
        await browser.close();
        console.log('Browser closed.');
    }
}

crawlGames();
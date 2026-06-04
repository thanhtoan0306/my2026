require('dotenv').config();
const { Telegraf } = require('telegraf');
const { exec } = require('child_process');
const path = require('path');
const fs = require('fs');

const botToken = process.env.TELEGRAM_BOT_TOKEN;
const allowedChatId = Number(process.env.ALLOWED_USER_ID);

if (!botToken || !allowedChatId) {
    console.error("❌ Thiếu TELEGRAM_BOT_TOKEN hoặc ALLOWED_USER_ID trong file .env!");
    process.exit(1);
}

const bot = new Telegraf(botToken);

let currentDir = process.cwd();

bot.use((ctx, next) => {
    if (ctx.chat && ctx.chat.id === allowedChatId) {
        return next();
    }
    console.warn(`⚠️ Cảnh báo: Người dùng có ID ${ctx.chat?.id} cố gắng truy cập bot.`);
    return ctx.reply("❌ Bạn không có quyền điều khiển PC này.");
});

bot.start((ctx) => {
    ctx.reply(`🤖 Bot đã sẵn sàng!\n📍 Thư mục hiện tại: \`${currentDir}\`\n\n💡 Bạn có thể dùng lệnh \`/cd <đường_dẫn>\` để chuyển thư mục.`);
});

bot.command('cd', async (ctx) => {
    const targetDir = ctx.message.text.split(' ').slice(1).join(' ').trim();

    if (!targetDir) {
        return ctx.reply(`📍 Thư mục hiện tại: \`${currentDir}\``);
    }

    const resolvedPath = path.resolve(currentDir, targetDir);

    if (fs.existsSync(resolvedPath) && fs.lstatSync(resolvedPath).isDirectory()) {
        currentDir = resolvedPath;
        await ctx.reply(`✅ Đã chuyển đến thư mục:\n\`${currentDir}\``);
    } else {
        await ctx.reply(`❌ Đường dẫn không tồn tại hoặc không phải là thư mục:\n\`${resolvedPath}\``);
    }
});

bot.on('text', async (ctx) => {
    const userPrompt = ctx.message.text.replace(/"/g, '\\"');
    await ctx.reply(`⏳ Đang chạy Gemini CLI tại:\n\`${currentDir}\`...`);

    const command = `gemini "${userPrompt}"`;

    exec(command, { encoding: 'utf-8', cwd: currentDir }, (error, stdout, stderr) => {
        let output = stdout || stderr;

        if (error) {
            return ctx.reply(`❌ Lỗi thực thi CLI: ${error.message}`);
        }

        if (!output.trim()) {
            output = "Thực thi xong nhưng không có phản hồi văn bản.";
        }

        if (output.length > 4000) {
            for (let i = 0; i < output.length; i += 4000) {
                ctx.reply(output.substring(i, i + 4000));
            }
        } else {
            ctx.reply(output);
        }
    });
});

bot.launch().then(() => {
    console.log("🚀 Bot đang chạy thành công trên PC...");
});

process.once('SIGINT', () => bot.stop('SIGINT'));
process.once('SIGTERM', () => bot.stop('SIGTERM'));

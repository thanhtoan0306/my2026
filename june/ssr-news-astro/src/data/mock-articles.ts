import type { Article } from '../types/news';

const TODAY = '2026-06-11';

export const MOCK_ARTICLES: Article[] = [
	{
		id: '1',
		slug: 'bitcoin-etf-inflows-record-june-2026',
		title: 'Bitcoin Tops $110K as Spot ETF Inflows Hit a New Daily Record',
		summary:
			'U.S. spot Bitcoin ETFs absorbed over $2.1B in a single session, pushing BTC past $110,000 as institutional demand accelerates ahead of summer.',
		body: `Spot Bitcoin exchange-traded funds listed in the United States recorded their largest single-day net inflow since launch, with BlackRock's IBIT and Fidelity's FBTC leading the charge.

Analysts point to pension fund allocation pilots and renewed macro hedging demand as drivers. On-chain data shows exchange reserves continuing to decline while long-term holder supply rises.

Traders are watching the $112K zone as the next psychological barrier. Options markets now price a 38% chance of BTC closing above $115K by month-end, up from 22% last week.

"Flows are telling the story," said one desk strategist. "Retail is participating through ETFs, but the size is unmistakably institutional."`,
		category: 'Bitcoin',
		author: 'Crypto Desk',
		publishedAt: `${TODAY}T08:15:00Z`,
		tags: ['bitcoin', 'etf', 'markets'],
	},
	{
		id: '2',
		slug: 'ethereum-pectra-mainnet-live',
		title: 'Ethereum Pectra Upgrade Goes Live, Unlocking Account Abstraction at Scale',
		summary:
			'The Pectra hard fork activated on mainnet today, bundling EIP-7702 and blob capacity tweaks that developers say will cut L2 fees and simplify wallet UX.',
		body: `Ethereum mainnet block 22,500,000 marked the activation of Pectra, the network's latest protocol upgrade combining Prague and Electra changes.

Validators began enforcing new rules around execution-layer triggers and increased blob throughput. Major rollups reported a 12–18% reduction in data posting costs within the first hours.

Wallet vendors highlighted EIP-7702, which lets externally owned accounts temporarily behave like smart contracts — a stepping stone toward native account abstraction without forcing users to migrate addresses.

The ETH/BTC ratio ticked up 1.4% on the news. Staking queues remain elevated as operators race to update client software.`,
		category: 'Ethereum',
		author: 'Layer-1 Team',
		publishedAt: `${TODAY}T11:42:00Z`,
		tags: ['ethereum', 'pectra', 'upgrade'],
	},
	{
		id: '3',
		slug: 'sec-stablecoin-framework-banks-2026',
		title: 'SEC Unveils Stablecoin Framework, Banks Rush to Issue USD-Backed Tokens',
		summary:
			'Regulators published final guidance on reserve audits and redemption rights, prompting several U.S. banks to announce tokenized deposit pilots within hours.',
		body: `The Securities and Exchange Commission released its long-awaited stablecoin compliance framework, clarifying how federally chartered banks may issue dollar-backed tokens under existing custody rules.

The guidance requires monthly attestations, same-day redemption for retail holders, and segregation of reserve assets. GENIUS Act provisions referenced in the release align federal and state oversight paths.

JPMorgan, Bank of New York Mellon, and two regional lenders disclosed pilot programs for blockchain-settled deposits targeting corporate treasury clients.

USDC and USDT collectively represent over $280B in circulation. Market participants expect competitive pressure on yield-bearing products as bank-issued tokens enter the field.

Crypto equities moved higher in afternoon trade, with exchange and custody names leading the sector.`,
		category: 'Regulation',
		author: 'Policy Wire',
		publishedAt: `${TODAY}T15:05:00Z`,
		tags: ['stablecoin', 'sec', 'regulation'],
	},
];

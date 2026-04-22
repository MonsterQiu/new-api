const DEFAULT_CURRENCY = 'USD';

const CURRENCY_SYMBOLS = {
  USD: '$',
  CNY: '¥',
  EUR: '€',
  GBP: '£',
  HKD: 'HK$',
  JPY: '¥',
  KRW: '₩',
  SGD: 'S$',
  AUD: 'A$',
  CAD: 'C$',
};

function normalizeCurrencyCode(currency) {
  const normalized = String(currency || DEFAULT_CURRENCY)
    .trim()
    .toUpperCase();
  return normalized || DEFAULT_CURRENCY;
}

export function getCurrencySymbolByCode(currency) {
  const normalized = normalizeCurrencyCode(currency);
  return CURRENCY_SYMBOLS[normalized] || normalized;
}

export function formatSubscriptionPaymentAmount(amount) {
  const numericAmount = Number(amount || 0);
  if (!Number.isFinite(numericAmount)) {
    return '0.00';
  }
  return numericAmount.toFixed(Number.isInteger(numericAmount) ? 0 : 2);
}

function buildPaymentOption(key, label, amount, currency) {
  const normalizedCurrency = normalizeCurrencyCode(currency);
  const formattedAmount = formatSubscriptionPaymentAmount(amount);
  return {
    key,
    label,
    amount: Number(amount || 0),
    currency: normalizedCurrency,
    symbol: getCurrencySymbolByCode(normalizedCurrency),
    formattedAmount,
    displayPrice: `${getCurrencySymbolByCode(normalizedCurrency)}${formattedAmount}`,
  };
}

function getCreemCurrency(planCurrency) {
  const quotaDisplayType = localStorage.getItem('quota_display_type') || 'USD';
  if (quotaDisplayType === 'CNY') {
    return 'CNY';
  }
  return normalizeCurrencyCode(planCurrency);
}

export function getSubscriptionPaymentOptions({
  plan,
  enableOnlineTopUp = false,
  enableStripeTopUp = false,
  enableCreemTopUp = false,
  epayMethods = [],
  epayLabel = 'Epay',
}) {
  const priceAmount = Number(plan?.price_amount || 0);
  const planCurrency = normalizeCurrencyCode(
    plan?.currency || DEFAULT_CURRENCY,
  );
  const options = [];

  if (enableStripeTopUp && plan?.stripe_price_id) {
    options.push(
      buildPaymentOption('stripe', 'Stripe', priceAmount, planCurrency),
    );
  }

  if (enableCreemTopUp && plan?.creem_product_id) {
    options.push(
      buildPaymentOption(
        'creem',
        'Creem',
        priceAmount,
        getCreemCurrency(planCurrency),
      ),
    );
  }

  if (enableOnlineTopUp && (epayMethods || []).length > 0) {
    options.push(buildPaymentOption('epay', epayLabel, priceAmount, 'CNY'));
  }

  if (options.length === 0) {
    options.push(buildPaymentOption('plan', '', priceAmount, planCurrency));
  }

  return options;
}

export function getGroupedSubscriptionPaymentOptions(input) {
  const groupedOptions = new Map();

  getSubscriptionPaymentOptions(input).forEach((option) => {
    const groupKey = `${option.currency}:${option.amount.toFixed(6)}`;
    const existing = groupedOptions.get(groupKey);

    if (existing) {
      if (option.label && !existing.labels.includes(option.label)) {
        existing.labels.push(option.label);
      }
      return;
    }

    groupedOptions.set(groupKey, {
      ...option,
      labels: option.label ? [option.label] : [],
    });
  });

  return Array.from(groupedOptions.values()).map((option) => ({
    ...option,
    label: option.labels.join(' / '),
  }));
}

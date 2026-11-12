# Mathematical Revenue Ranking Matrix

Traditional product management tools count raw feedback frequency. This approach often over-indexes on minor bugs submitted by free or low-tier users while overlooking enterprise blocking requirements demanded by high-value accounts.

feedbackloom ranks features by the actual dollars at stake using the **Weighted Revenue Score ($WRS$)**.

## 1. Total ARR Formulation

For an opportunity cluster $K$ containing feedback items $\{f_1, f_2, \dots, f_n\}$ originating from customer accounts $A(f_i)$:

$$\text{Total ARR}(K) = \sum_{a \in \text{UniqueAccounts}(K)} \text{ARR}(a)$$

Deduplication by unique account ensures that an account submitting multiple tickets regarding the same problem contributes its ARR only once.

## 2. Weighted Revenue Score (WRS)

To prioritize features demanded by accounts nearing contract renewal or showing high churn indicators, the Weighted Revenue Score is defined as:

$$WRS(K) = \sum_{a \in \text{UniqueAccounts}(K)} \left[ \text{ARR}(a) \times \left( 1 + \frac{1}{\max(1, \text{DaysToRenewal}(a))} \right) \times \text{ChurnRisk}(a) \right]$$

### Variables:
- **$\text{ARR}(a)$**: Annual Recurring Revenue of account $a$ in USD from Stripe or CRM records.
- **$\text{DaysToRenewal}(a)$**: Number of calendar days until the account contract or subscription renewal date.
- **$\text{ChurnRisk}(a)$**: Account churn probability score ranging from 0.0 (fully retained) to 1.0 (imminent cancellation).

## 3. Tier Segmentation

Each opportunity cluster tracks revenue distribution across three tiers:
- **Enterprise ARR**: Accounts with $\text{ARR} \ge \$50,000$.
- **Growth ARR**: Accounts with $\$10,000 \le \text{ARR} < \$50,000$.
- **Starter ARR**: Accounts with $\text{ARR} < \$10,000$.

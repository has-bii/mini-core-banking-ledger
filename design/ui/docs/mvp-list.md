# Still Bank — MVP Feature List

## Design Approach

Design the UI to mimic a mobile app, using app-style screens, navigation, and interactions—not a responsive website adapted for a small screen. The frontend implementation does not imply a native app build.

## Scope

Still Bank's MVP is an interactive, mobile-app-style frontend that demonstrates complete banking journeys using simulated data and transactions. Backend implementation and real-money movement are outside this document's scope.

This list defines product features, not technical implementation details, priorities, or acceptance criteria. Where wallet and savings behavior differs from `problem-and-users.md`, this document reflects the newer decisions.

## Authentication

- Register an account with full name, email, and password.
- Verify registration through an email OTP.
- Sign in with email and password after verification.
- Recover a forgotten password by entering the account email, verifying an OTP, and setting a new password. Sign in again with the new password.
- Change the password while signed in.

## Device Enrollment and Unlock

- Keep account registration separate from enrolling an account on a device.
- After signing in, optionally enroll the device for quick unlock using an account PIN or passkey.
- Let returning users unlock an enrolled account without entering email and password again.
- Support passkey unlock through the device's available method, such as biometrics or its device PIN. The Still Bank account PIN is separate from the device PIN.
- Add or change the unlock method and remove a registered passkey.
- Reset a forgotten account PIN through email OTP verification.
- Lock the account while retaining device enrollment, returning to the quick-unlock screen.
- Log out by removing the account's enrollment from that device. Require email/password sign-in to enroll it again.
- Offer “Use another account” on the returning-user unlock screen.

## Wallet Management

- Automatically provide exactly one main wallet per user.
- Open sub wallets and savings wallets.
- Give each wallet a unique wallet number.
- View and rename wallets, including the main wallet.
- Close sub wallets and savings wallets, automatically moving their remaining balances to the main wallet.
- Prevent closing the main wallet.
- Require removal of an unfinished savings target before closing its wallet.
- Cancel a savings wallet's budgeting schedule when the wallet closes.

## Savings Targets

- Set a savings target amount with or without a deadline.
- View progress toward the target.
- Edit or remove a target.
- Block outgoing money moves and transfers while a target is unfinished.
- Unlock outgoing money when the target amount is reached or the target is removed.
- Treat the deadline as a progress reference, not a condition that unlocks money.
- Continue accepting incoming money while outgoing money is locked.

## Scheduled Budgeting

- Optionally configure budgeting for a savings wallet with an amount, source wallet, period, and finish date.
- Select a main or sub wallet as the source.
- Support daily, weekly, and monthly periods.
- Display a schedule description and a **Move money** button inside the savings wallet.
- Require the user to initiate each contribution; do not deduct money automatically.
- Advance the schedule by one period after each successful contribution.
- Allow early contributions: after contributing, the user can contribute again and advance the schedule another period.
- Handle overdue contributions one installment per click.
- Keep the schedule unchanged when a contribution fails.
- Edit or cancel a budgeting schedule.
- End budgeting after its finish date.
- Keep budgeting and targets independent: a savings wallet may have either, both, or neither.

## Top-up

- Enter an amount and select any open destination wallet.
- Review and confirm the top-up.
- Show a success result only.
- Allow top-ups into target-locked savings wallets.
- Simulate funding without selecting a payment method or integrating a payment gateway.

## Move and Transfer Money

- Move money between the user's own wallets.
- Transfer money to another Still Bank user's wallet using its wallet number.
- Display the recipient's name before confirmation.
- Allow main, sub, and savings wallets to receive transfers.
- Select the source, destination, and amount; review the details before confirming.
- Show a success or failure result.
- Block transactions with insufficient funds and outgoing transactions from target-locked savings wallets.
- Do not charge transfer fees.

## Balances and Activity

- Display the total balance and individual wallet balances.
- Keep daily spending, reserved money, and savings distinguishable.
- Show transaction history across all wallets and within each wallet.
- Show transaction details and their effect on wallet balances.
- Distinguish moves between the user's wallets from transfers to other users.

## Settings and Currency

- Provide settings for password, account PIN, and passkey management.
- Provide account locking and logout actions.
- Use IDR as the only currency.

## Out of Scope

- Backend implementation and real-money movement.
- Live banking integrations, payment gateways, and external bank transfers.
- Withdrawals, bill payments, and QR payments.
- Cards or separate bank accounts per wallet; wallet numbers identify wallets within Still Bank only.
- Loans, investments, interest-bearing savings products, and business banking.
- Fully automatic savings deductions.
- Scheduled transfers and saved recipients.
- Profile editing, notifications, and device-management lists.
- Desktop-specific layouts.

# Still Bank — Problem and Users

## Product context

Still Bank is a mobile-first banking experience focused on simple money organization, thoughtful visual design, and clear everyday interactions.

It is a learning and portfolio project, not a production banking service. The interface should feel like a complete banking product, without demo labels. This presentation choice does not imply real banking capabilities.

The frontend will use React, with mobile UI and usability taking priority.

## Who it is for

The primary user is an individual who banks on their phone and wants to separate everyday spending, reserved money, and savings goals without complicated budgeting.

The project creator is the initial reference user. The audience is defined by needs rather than assumed age, income, or occupation.

### Core needs

- Know how much money is available for everyday spending.
- Keep money for different purposes separate and easy to understand.
- Set money aside for savings goals and see progress.
- Make transfers and clearly understand their outcome.
- Review transactions and understand their effect on balances.
- Complete everyday tasks through a simple, well-designed mobile interface.

Business banking and bank employee workflows are not the initial focus.

## The problem

> I need a simple way to organize everyday money and savings, so I know what I can spend without taking money intended for something else.

The working hypothesis is that keeping money mentally allocated—or tracking those allocations in another app—makes it harder to distinguish spendable money from money reserved for other purposes.

Still Bank aims to make those distinctions visible and manageable within the banking experience itself. Simplicity means helping users understand their money and complete everyday tasks, not merely reducing the number of elements on a screen.

## Initial product scope

### Purpose-based wallets

Wallets are separate money containers within one banking experience, each with its own balance. Examples include Daily Spending, Bills, and Personal.

Users can create wallets and move money between them. Money reserved in another wallet remains distinct from daily spending money. A Bills wallet organizes money; it does not imply bill-payment functionality.

Separate bank accounts, cards, and account numbers per wallet are outside the initial definition.

### Goal-based savings

Users can create a savings goal, set a target amount, contribute manually, and understand progress toward that target.

Interest-bearing savings products and investments are outside the initial scope.

### Everyday banking essentials

- Check balances.
- Transfer money.
- Review transaction history.

Bill payments, QR payments, loans, and investments are outside the initial scope. Desktop is not the primary design target.

## What success looks like

Success means users can complete key tasks correctly and confidently—not that the project reaches a particular number of sign-ups or generates revenue.

| Outcome | Usability criterion |
| --- | --- |
| Understand available money | The user can distinguish daily spending money from money reserved in other wallets or savings. |
| Organize money | The user can create a wallet and move money into it without assistance. |
| Save toward a goal | The user can create a savings goal, contribute, and understand progress. |
| Transfer confidently | The user can complete a transfer and clearly understand its outcome. |
| Understand activity | The user can find a transaction and understand its effect on their balance. |

Visual design succeeds when it supports these outcomes through clear hierarchy, understandable information, and straightforward interactions. Appearance alone is not sufficient.

## Evidence and validation status

This definition is an initial product hypothesis based on the project creator's needs. It is not supported by user interviews or usability testing yet.

The success criteria above are targets for future evaluation, not achieved results. Future testing should examine whether users understand the separation between spending money, reserved funds, and savings, and whether they can complete the key tasks without assistance.

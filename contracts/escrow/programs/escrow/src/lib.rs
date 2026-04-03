use anchor_lang::prelude::*;
use anchor_lang::system_program;

declare_id!("EscrowProgramIDPlaceholder11111111111111111");

#[program]
pub mod screener_escrow {
    use super::*;

    /// Buyer deposits SOL into a PDA-controlled escrow account for a dataset purchase.
    pub fn deposit(ctx: Context<Deposit>, amount: u64, dataset_id: [u8; 16]) -> Result<()> {
        require!(amount > 0, EscrowError::InvalidAmount);

        // Transfer SOL from buyer to escrow vault PDA
        system_program::transfer(
            CpiContext::new(
                ctx.accounts.system_program.to_account_info(),
                system_program::Transfer {
                    from: ctx.accounts.buyer.to_account_info(),
                    to: ctx.accounts.vault.to_account_info(),
                },
            ),
            amount,
        )?;

        // Initialize escrow state
        let escrow = &mut ctx.accounts.escrow_state;
        escrow.buyer = ctx.accounts.buyer.key();
        escrow.authority = ctx.accounts.authority.key();
        escrow.dataset_id = dataset_id;
        escrow.amount = amount;
        escrow.released = 0;
        escrow.status = EscrowStatus::Active;
        escrow.bump = ctx.bumps.vault;

        emit!(DepositEvent {
            buyer: escrow.buyer,
            dataset_id,
            amount,
            escrow: ctx.accounts.escrow_state.key(),
            vault: ctx.accounts.vault.key(),
        });

        Ok(())
    }

    /// Authority releases a portion of escrowed SOL to a seller.
    /// Can be called multiple times (once per seller) until the escrow is drained.
    pub fn release(ctx: Context<Release>, amount: u64) -> Result<()> {
        let escrow = &mut ctx.accounts.escrow_state;

        require!(escrow.status == EscrowStatus::Active, EscrowError::EscrowNotActive);
        require!(amount > 0, EscrowError::InvalidAmount);
        require!(
            escrow.released.checked_add(amount).unwrap() <= escrow.amount,
            EscrowError::InsufficientEscrowBalance
        );

        // Transfer SOL from vault PDA to seller
        let buyer_key = escrow.buyer;
        let dataset_id = escrow.dataset_id;
        let bump = escrow.bump;

        let seeds = &[
            b"vault",
            buyer_key.as_ref(),
            dataset_id.as_ref(),
            &[bump],
        ];
        let signer_seeds = &[&seeds[..]];

        **ctx.accounts.vault.to_account_info().try_borrow_mut_lamports()? -= amount;
        **ctx.accounts.seller.to_account_info().try_borrow_mut_lamports()? += amount;

        escrow.released += amount;

        // Mark completed if fully released
        if escrow.released >= escrow.amount {
            escrow.status = EscrowStatus::Completed;
        }

        emit!(ReleaseEvent {
            escrow: ctx.accounts.escrow_state.key(),
            seller: ctx.accounts.seller.key(),
            amount,
            total_released: escrow.released,
        });

        Ok(())
    }

    /// Authority refunds all remaining SOL back to the buyer.
    /// Used if the purchase fails or is cancelled.
    pub fn refund(ctx: Context<Refund>) -> Result<()> {
        let escrow = &mut ctx.accounts.escrow_state;

        require!(escrow.status == EscrowStatus::Active, EscrowError::EscrowNotActive);

        let remaining = escrow.amount - escrow.released;
        require!(remaining > 0, EscrowError::NothingToRefund);

        let buyer_key = escrow.buyer;
        let dataset_id = escrow.dataset_id;
        let bump = escrow.bump;

        let seeds = &[
            b"vault",
            buyer_key.as_ref(),
            dataset_id.as_ref(),
            &[bump],
        ];
        let signer_seeds = &[&seeds[..]];

        **ctx.accounts.vault.to_account_info().try_borrow_mut_lamports()? -= remaining;
        **ctx.accounts.buyer.to_account_info().try_borrow_mut_lamports()? += remaining;

        escrow.status = EscrowStatus::Refunded;

        emit!(RefundEvent {
            escrow: ctx.accounts.escrow_state.key(),
            buyer: ctx.accounts.buyer.key(),
            amount: remaining,
        });

        Ok(())
    }
}

// ── Accounts ──────────────────────────────────────────────────────────────────

#[derive(Accounts)]
#[instruction(amount: u64, dataset_id: [u8; 16])]
pub struct Deposit<'info> {
    #[account(mut)]
    pub buyer: Signer<'info>,

    /// The server backend wallet, set as the release/refund authority.
    /// CHECK: This is just stored as a pubkey in the escrow state.
    pub authority: UncheckedAccount<'info>,

    #[account(
        init,
        payer = buyer,
        space = 8 + EscrowState::INIT_SPACE,
        seeds = [b"escrow", buyer.key().as_ref(), dataset_id.as_ref()],
        bump,
    )]
    pub escrow_state: Account<'info, EscrowState>,

    #[account(
        mut,
        seeds = [b"vault", buyer.key().as_ref(), dataset_id.as_ref()],
        bump,
    )]
    /// CHECK: PDA-controlled vault that holds the escrowed SOL.
    pub vault: UncheckedAccount<'info>,

    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct Release<'info> {
    /// Only the authority can release funds.
    #[account(
        constraint = authority.key() == escrow_state.authority @ EscrowError::Unauthorized
    )]
    pub authority: Signer<'info>,

    #[account(
        mut,
        seeds = [b"escrow", escrow_state.buyer.as_ref(), escrow_state.dataset_id.as_ref()],
        bump,
    )]
    pub escrow_state: Account<'info, EscrowState>,

    #[account(
        mut,
        seeds = [b"vault", escrow_state.buyer.as_ref(), escrow_state.dataset_id.as_ref()],
        bump = escrow_state.bump,
    )]
    /// CHECK: PDA vault holding the escrowed SOL.
    pub vault: UncheckedAccount<'info>,

    /// CHECK: The seller receiving the payout. Validated by the backend.
    #[account(mut)]
    pub seller: UncheckedAccount<'info>,

    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct Refund<'info> {
    /// Only the authority can refund.
    #[account(
        constraint = authority.key() == escrow_state.authority @ EscrowError::Unauthorized
    )]
    pub authority: Signer<'info>,

    #[account(
        mut,
        seeds = [b"escrow", escrow_state.buyer.as_ref(), escrow_state.dataset_id.as_ref()],
        bump,
    )]
    pub escrow_state: Account<'info, EscrowState>,

    #[account(
        mut,
        seeds = [b"vault", escrow_state.buyer.as_ref(), escrow_state.dataset_id.as_ref()],
        bump = escrow_state.bump,
    )]
    /// CHECK: PDA vault holding the escrowed SOL.
    pub vault: UncheckedAccount<'info>,

    /// CHECK: The original buyer receiving the refund.
    #[account(
        mut,
        constraint = buyer.key() == escrow_state.buyer @ EscrowError::Unauthorized
    )]
    pub buyer: UncheckedAccount<'info>,

    pub system_program: Program<'info, System>,
}

// ── State ─────────────────────────────────────────────────────────────────────

#[account]
#[derive(InitSpace)]
pub struct EscrowState {
    pub buyer: Pubkey,         // 32
    pub authority: Pubkey,     // 32
    pub dataset_id: [u8; 16], // 16
    pub amount: u64,           // 8
    pub released: u64,         // 8
    pub status: EscrowStatus,  // 1
    pub bump: u8,              // 1
}

#[derive(AnchorSerialize, AnchorDeserialize, Clone, Copy, PartialEq, Eq, InitSpace)]
pub enum EscrowStatus {
    Active,
    Completed,
    Refunded,
}

// ── Events ────────────────────────────────────────────────────────────────────

#[event]
pub struct DepositEvent {
    pub buyer: Pubkey,
    pub dataset_id: [u8; 16],
    pub amount: u64,
    pub escrow: Pubkey,
    pub vault: Pubkey,
}

#[event]
pub struct ReleaseEvent {
    pub escrow: Pubkey,
    pub seller: Pubkey,
    pub amount: u64,
    pub total_released: u64,
}

#[event]
pub struct RefundEvent {
    pub escrow: Pubkey,
    pub buyer: Pubkey,
    pub amount: u64,
}

// ── Errors ────────────────────────────────────────────────────────────────────

#[error_code]
pub enum EscrowError {
    #[msg("Only the authority can perform this action")]
    Unauthorized,
    #[msg("Amount must be greater than zero")]
    InvalidAmount,
    #[msg("Escrow is not in active state")]
    EscrowNotActive,
    #[msg("Release amount exceeds remaining escrow balance")]
    InsufficientEscrowBalance,
    #[msg("No funds remaining to refund")]
    NothingToRefund,
}

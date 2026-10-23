import { forwardRef, type ButtonHTMLAttributes, type ReactNode } from 'react'
import { motion, type HTMLMotionProps } from 'framer-motion'
import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/cn'
import { Tooltip } from './Tooltip'

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger' | 'subtle'
type Size = 'sm' | 'md' | 'lg'

const variants: Record<Variant, string> = {
  primary: 'bg-accent-gradient text-accent-fg shadow-glow hover:brightness-110 disabled:shadow-none',
  secondary: 'bg-elevated text-fg border hover:bg-line/[0.06] dark:bg-line/[0.05] dark:hover:bg-line/[0.09]',
  ghost: 'text-muted hover:text-fg hover:bg-line/[0.06]',
  subtle: 'bg-accent/10 text-accent-ink hover:bg-accent/15',
  danger: 'bg-danger/10 text-danger hover:bg-danger/15',
}

const sizes: Record<Size, string> = {
  sm: 'h-8 px-3 text-xs gap-1.5 rounded-lg',
  md: 'h-9 px-3.5 text-[13px] gap-2 rounded-xl',
  lg: 'h-11 px-5 text-sm gap-2 rounded-xl font-semibold',
}

export interface ButtonProps extends Omit<HTMLMotionProps<'button'>, 'children'> {
  variant?: Variant
  size?: Size
  icon?: ReactNode
  iconEnd?: ReactNode
  loading?: boolean
  children?: ReactNode
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { variant = 'secondary', size = 'md', icon, iconEnd, loading, className, children, disabled, ...rest },
  ref,
) {
  return (
    <motion.button
      ref={ref}
      whileTap={disabled || loading ? undefined : { scale: 0.97 }}
      transition={{ type: 'spring', stiffness: 600, damping: 30 }}
      className={cn(
        'no-drag inline-flex select-none items-center justify-center whitespace-nowrap font-medium transition-[background-color,color,filter,box-shadow] duration-150',
        'disabled:cursor-not-allowed disabled:opacity-50',
        variants[variant],
        sizes[size],
        className,
      )}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      {...rest}
    >
      {loading ? <Loader2 className="size-4 animate-spin" aria-hidden /> : icon}
      {children}
      {iconEnd}
    </motion.button>
  )
})

interface IconButtonProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'children'> {
  label: string
  children: ReactNode
  size?: 'sm' | 'md'
  tone?: 'default' | 'danger'
  tooltip?: boolean
}

/** Square icon-only button. `label` is used for aria-label and the tooltip. */
export const IconButton = forwardRef<HTMLButtonElement, IconButtonProps>(function IconButton(
  { label, children, className, size = 'md', tone = 'default', tooltip = true, ...rest },
  ref,
) {
  const btn = (
    <button
      ref={ref}
      type="button"
      aria-label={label}
      className={cn(
        'no-drag inline-flex shrink-0 items-center justify-center rounded-lg text-muted transition-colors duration-150',
        'hover:bg-line/[0.07] hover:text-fg active:scale-95 disabled:pointer-events-none disabled:opacity-40',
        tone === 'danger' && 'hover:bg-danger/10 hover:text-danger',
        size === 'sm' ? 'size-7' : 'size-8',
        className,
      )}
      {...rest}
    >
      {children}
    </button>
  )
  return tooltip ? <Tooltip content={label}>{btn}</Tooltip> : btn
})

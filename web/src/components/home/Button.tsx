import React from 'react'
import { ButtonProps } from '../../types'
import Link from 'next/link'

const Button: React.FC<ButtonProps> = ({
  children,
  onClick,
  variant = 'primary',
  size = 'md',
  className = '',
  href,
  type = 'button',
  icon,
}) => {
  const baseStyles =
    'inline-flex items-center justify-center font-semibold rounded-2xl focus:outline-none focus:ring-2 focus:ring-offset-2 transition-colors duration-150 ease-in-out'

  let variantStyles = ''
  switch (variant) {
    case 'primary':
      variantStyles = 'bg-blue-600 text-white hover:bg-blue-700 focus:ring-blue-500'
      break
    case 'secondary':
      variantStyles = 'bg-slate-200 text-slate-700 hover:bg-slate-300 focus:ring-slate-400'
      break
    case 'outline':
      variantStyles = 'border border-blue-600 text-blue-600 hover:bg-blue-50 focus:ring-blue-500'
      break
  }

  let sizeStyles = ''
  switch (size) {
    case 'sm':
      sizeStyles = 'px-3 py-1.5 text-sm'
      break
    case 'md':
      sizeStyles = 'px-6 py-3 text-base' // Increased padding for better feel from p-4 or p-6
      break
    case 'lg':
      sizeStyles = 'px-8 py-3.5 text-lg' // Increased padding
      break
  }

  const combinedClassName = `${baseStyles} ${variantStyles} ${sizeStyles} ${className}`

  if (href) {
    return (
      <Link href={href} className={combinedClassName} rel="noopener noreferrer">
        {icon && <span className="mr-2 -ml-1 h-5 w-5">{icon}</span>}
        {children}
      </Link>
    )
  }

  return (
    <button type={type} onClick={onClick} className={combinedClassName}>
      {icon && <span className="mr-2 -ml-1 h-5 w-5">{icon}</span>}
      {children}
    </button>
  )
}

export default Button

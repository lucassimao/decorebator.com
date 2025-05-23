
import React from 'react';

export interface Feature {
  // Fix: Specify IconProps for the icon ReactElement to allow className in React.cloneElement
  icon: React.ReactElement<IconProps>;
  title: string;
  description: string;
}

export interface ButtonProps {
  children: React.ReactNode;
  onClick?: () => void;
  variant?: 'primary' | 'secondary' | 'outline';
  size?: 'sm' | 'md' | 'lg';
  className?: string;
  href?: string;
  type?: 'button' | 'submit' | 'reset';
  icon?: React.ReactElement;
}

export interface IconProps {
  className?: string;
}
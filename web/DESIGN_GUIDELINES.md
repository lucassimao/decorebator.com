# Decorebator Web Landing Page - Design Guidelines

This document provides comprehensive design guidelines for the Decorebator web landing page, including visual design, component architecture, animations, and implementation patterns.

## Table of Contents

1. [Visual Design System](#visual-design-system)
2. [Component Architecture](#component-architecture)
3. [Animation Guidelines](#animation-guidelines)
4. [Typography & Color System](#typography--color-system)
5. [Layout & Spacing](#layout--spacing)
6. [Component Library](#component-library)
7. [Responsive Design](#responsive-design)
8. [Technical Implementation](#technical-implementation)
9. [Brand Guidelines](#brand-guidelines)
10. [Future Development Patterns](#future-development-patterns)

---

## Visual Design System

### Color Palette

**Primary Colors:**

- `#FF7B54` - Primary orange (main brand color)
- `#FFD700` - Secondary gold/yellow (accent color)
- `#FDF6E3` - Background cream (main background)

**Text Colors:**

- `#2D3436` - Primary text (dark gray, almost black)
- `#636E72` - Secondary text (medium gray)
- `#ffffff` - White text (for dark backgrounds)

**Semantic Colors:**

- `#4CAF50` - Success green
- `#9C27B0` - Purple accent
- `#ff6347` - Hover orange (darker primary)

**Background Gradients:**

```css
/* Primary Button Gradient */
background: linear-gradient(to right, #ff7b54, #ff6347);

/* Hero Text Gradient */
background: linear-gradient(270deg, #ff7b54, #ffd700, #ff7b54);

/* Feature Card Icon Gradient */
background: linear-gradient(to bottom right, #ff7b54, #ff6347);
```

### Visual Hierarchy

**Page Structure:**

1. Fixed header with glassmorphism effect
2. Hero section with animated background elements
3. Features section with white background
4. Additional sections following alternating background pattern
5. Footer with brand color

**Z-Index Layers:**

- `z-50`: Fixed header
- `z-10`: Main content sections
- `z-0`: Background elements (non-interactive)

---

## Component Architecture

### Layout Components

**File Structure:**

```
src/components/layout/
├── PageLayout.tsx          # Main page wrapper
├── Header.tsx              # Fixed navigation header
└── BackgroundElements.tsx  # Animated background decorations
```

**PageLayout Component:**

- Purpose: Consistent page wrapper for all pages
- Props: `children: React.ReactNode`, `className?: string`
- Features: Background elements, header, main content area
- Usage: Wrap all page content

**Header Component:**

- Purpose: Fixed navigation with scroll effects
- Features: Logo, navigation links, mobile menu, CTA button
- States: Scroll detection, mobile menu toggle
- Responsive: Desktop nav hidden on mobile, hamburger menu shown

**BackgroundElements Component:**

- Purpose: Animated floating elements
- Features: Three floating circles with different animations
- Position: Fixed, behind all content (pointer-events: none)

### Content Components

**File Structure:**

```
src/components/home/
├── EnhancedHeroSection.tsx    # Main hero with phone mockup
├── FeaturesSection.tsx        # Feature grid container
├── FeatureCard.tsx           # Individual feature card
├── HowItWorksSection.tsx     # Process explanation
├── PricingSection.tsx        # Subscription plans
├── SocialProofSection.tsx    # Testimonials/reviews
├── SubscriptionCalloutSection.tsx # CTA section
├── EmailCaptureForm.tsx      # Newsletter signup
├── FooterSection.tsx         # Site footer
└── icons.tsx                 # Icon components
```

---

## Animation Guidelines

### Animation File Location

All animations are defined in: `src/styles/animations.css`

### Core Animations

**Float Animation (6s cycle):**

```css
@keyframes float {
  0%,
  100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-20px);
  }
}
```

- Usage: Background elements, floating decorations
- Duration: 6 seconds
- Easing: ease-in-out infinite

**Pulse Glow (3s cycle):**

```css
@keyframes pulse-glow {
  0%,
  100% {
    opacity: 0.5;
    transform: scale(1);
  }
  50% {
    opacity: 0.8;
    transform: scale(1.1);
  }
}
```

- Usage: Accent elements, call-to-action highlights
- Duration: 3 seconds
- Easing: ease-in-out infinite

**Slide-in Animations (0.8s):**

```css
/* slide-in-left, slide-in-right, slide-in-up */
```

- Usage: Page entry animations
- Duration: 0.8 seconds
- Easing: ease-out forwards

**Gradient Animation (8s cycle):**

```css
@keyframes gradient-shift {
  0% {
    background-position: 0% 50%;
  }
  50% {
    background-position: 100% 50%;
  }
  100% {
    background-position: 0% 50%;
  }
}
```

- Usage: Hero title text gradient
- Duration: 8 seconds
- Background size: 200% 200%

**Word Rotation (12s cycle):**

```css
@keyframes word-rotate {
  0%,
  25% {
    opacity: 1;
    transform: translateY(0);
  }
  33%,
  100% {
    opacity: 0;
    transform: translateY(-20px);
  }
}
```

- Usage: Hero title rotating words
- Duration: 12 seconds total (3s per word)
- Delays: 0s, 3s, 6s, 9s for 4 words

### Interaction Animations

**Hover Effects:**

- Scale: `transform: scale(1.05)` for buttons
- Translate: `transform: translateY(-2px)` for cards
- Color transitions: 300ms duration
- Shadow enhancement: `hover:shadow-2xl`

**3D Card Effect:**

```css
.card-3d {
  transform-style: preserve-3d;
  transition: transform 0.6s;
}
.card-3d:hover {
  transform: rotateY(10deg) rotateX(10deg);
}
```

### Animation Timing Guidelines

**Fast Interactions (< 300ms):**

- Button hover states
- Link color changes
- Small scale transformations

**Medium Interactions (300ms - 600ms):**

- Card hover effects
- Menu transitions
- Icon transformations

**Slow Animations (> 600ms):**

- Page entry animations
- Background elements
- Gradient shifts

---

## Typography & Color System

### Font System

**Font Family:**

- Primary: Geist Sans (via Next.js)
- Fallback: Arial, Helvetica, sans-serif
- Monospace: Geist Mono (for code elements)

**Font Weights:**

- Regular: 400
- Medium: 500
- Semibold: 600
- Bold: 700

**Font Sizes (Tailwind Classes):**

```css
/* Hero Titles */
.text-5xl lg:text-6xl  /* 48px/60px -> 60px/72px */

/* Section Titles */
.text-4xl lg:text-5xl  /* 36px/40px -> 48px/56px */

/* Subsection Titles */
.text-xl              /* 20px/28px */

/* Body Text */
.text-lg              /* 18px/28px */

/* Small Text */
.text-sm              /* 14px/20px */
```

### Text Color Applications

**Headers & Titles:**

- Primary: `text-[#2D3436]` (dark gray)
- Gradient: `bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent`

**Body Text:**

- Primary: `text-[#636E72]` (medium gray)
- Light backgrounds: `text-[#2D3436]`
- Dark backgrounds: `text-white`

**Interactive Elements:**

- Links: `text-[#636E72] hover:text-[#FF7B54]`
- Buttons: `text-white` on colored backgrounds

---

## Layout & Spacing

### Container System

**Max Width Containers:**

```css
.max-w-7xl        /* 1280px - Main content wrapper */
.max-w-4xl        /* 896px - Text-heavy pages (privacy, terms) */
.max-w-3xl        /* 768px - Hero subtitle */
```

**Padding System:**

```css
/* Horizontal Padding */
.px-4 sm:px-6 lg:px-8    /* Responsive horizontal padding */

/* Vertical Padding */
.py-20                   /* Section spacing (80px top/bottom) */
.pt-24 pb-20            /* Page content (96px top, 80px bottom) */
```

### Grid Systems

**Feature Grid:**

```css
.grid .grid-cols-1 .md:grid-cols-2 .lg:grid-cols-3 .gap-8
```

**Hero Grid:**

```css
.grid .lg:grid-cols-2 .gap-12 .items-center
```

### Spacing Scale

**Margin/Padding Values:**

- `space-y-8`: 32px vertical spacing (hero content)
- `gap-8`: 32px grid gaps
- `gap-4`: 16px small gaps
- `mb-16`: 64px section bottom margin
- `mb-12`: 48px subsection spacing
- `mb-4`: 16px paragraph spacing

---

## Component Library

### Button Components

**Primary Button:**

```tsx
<button className="group bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-8 py-4 rounded-full font-semibold text-lg hover:shadow-2xl transform hover:scale-105 transition-all duration-300 flex items-center justify-center">
```

**Secondary Button:**

```tsx
<button className="group bg-white/80 backdrop-blur px-8 py-4 rounded-full font-semibold text-lg border-2 border-gray-200 hover:border-[#FF7B54] transition-all duration-300 flex items-center justify-center">
```

**Header CTA Button:**

```tsx
<button className="bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-6 py-2.5 rounded-full font-semibold hover:shadow-lg transform hover:scale-105 transition-all duration-300">
```

### Card Components

**Feature Card Structure:**

```tsx
<div className="group card-3d transform rounded-3xl border border-gray-100 bg-white/80 p-8 shadow-xl backdrop-blur transition-all duration-500 hover:-translate-y-2 hover:shadow-2xl">
  <div className="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-[#FF7B54] to-orange-600 p-4 text-white transition-transform duration-300 group-hover:scale-110">
    {/* Icon */}
  </div>
  <h3 className="mb-4 text-xl font-bold text-[#2D3436] transition-colors duration-300 group-hover:text-[#FF7B54]">
    {/* Title */}
  </h3>
  <p className="leading-relaxed text-[#636E72]">{/* Description */}</p>
</div>
```

### Glassmorphism Effect

**Glass Class:**

```css
.glass {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
}
```

**Usage:**

- Header background
- Modal overlays
- Floating elements
- Card overlays

---

## Responsive Design

### Breakpoint System (Tailwind)

```css
sm: 640px   /* Small screens */
md: 768px   /* Medium screens */
lg: 1024px  /* Large screens */
xl: 1280px  /* Extra large screens */
```

### Mobile-First Patterns

**Typography Scaling:**

```css
/* Mobile -> Desktop */
text-4xl lg:text-5xl    /* 36px -> 48px */
text-5xl lg:text-6xl    /* 48px -> 60px */
```

**Layout Changes:**

```css
/* Mobile stacked -> Desktop side-by-side */
.grid .lg:grid-cols-2

/* Mobile 1 column -> Desktop 3 columns */
.grid .grid-cols-1 .md:grid-cols-2 .lg:grid-cols-3
```

**Spacing Adjustments:**

```css
/* Responsive padding */
.px-4 .sm:px-6 .lg:px-8

/* Responsive gaps */
.gap-6 .md:gap-8
```

### Mobile Specific Components

**Mobile Menu:**

- Hidden by default (`hidden md:hidden`)
- Toggle state managed in Header component
- Backdrop blur effect
- Slide down animation

**Hero Phone Mockup:**

- Responsive sizing: `w-72 lg:w-80`
- Rotation effect on hover
- Proper touch targets on mobile

---

## Technical Implementation

### File Organization

```
web/src/
├── app/
│   ├── layout.tsx              # Root layout with fonts
│   ├── page.tsx                # Home page composition
│   ├── privacy/page.tsx        # Privacy policy with layout
│   ├── terms/page.tsx          # Terms of service with layout
│   └── globals.css             # Global styles + animation imports
├── components/
│   ├── layout/
│   │   ├── PageLayout.tsx      # Main layout wrapper
│   │   ├── Header.tsx          # Navigation header
│   │   └── BackgroundElements.tsx
│   └── home/
│       ├── EnhancedHeroSection.tsx
│       ├── FeaturesSection.tsx
│       ├── FeatureCard.tsx
│       └── [...other sections]
├── styles/
│   └── animations.css          # All animation definitions
└── types.ts                    # TypeScript interfaces
```

### Key Dependencies

**Required in package.json:**

```json
{
  "dependencies": {
    "next": "15.3.2",
    "react": "^19.0.0",
    "react-dom": "^19.0.0"
  },
  "devDependencies": {
    "tailwindcss": "^4",
    "typescript": "^5"
  }
}
```

**External Resources:**

- Font Awesome 6.4.0 (via CDN in layout.tsx)
- App Store badges (via Wikimedia)
- Google Play badges (via Wikimedia)

### Component Props Patterns

**Layout Components:**

```tsx
interface PageLayoutProps {
  children: React.ReactNode
  className?: string
}
```

**Feature Components:**

```tsx
interface FeatureCardProps {
  feature: Feature
  index?: number // For staggered animations
}
```

**Section Components:**

```tsx
interface FeaturesSectionProps {
  features: Feature[]
}
```

### State Management Patterns

**Hero Word Rotation:**

```tsx
const [currentWordIndex, setCurrentWordIndex] = useState(0)
const words = ['AI Intelligence', 'Spaced Repetition', 'Visual Learning', 'Smart Quizzes']

useEffect(() => {
  const interval = setInterval(() => {
    setCurrentWordIndex((prev) => (prev + 1) % words.length)
  }, 3000)
  return () => clearInterval(interval)
}, [words.length])
```

**Header Scroll Detection:**

```tsx
const [isScrolled, setIsScrolled] = useState(false)

useEffect(() => {
  const handleScroll = () => {
    setIsScrolled(window.scrollY > 0)
  }
  window.addEventListener('scroll', handleScroll)
  return () => window.removeEventListener('scroll', handleScroll)
}, [])
```

---

## Brand Guidelines

### Logo Usage

**Logo Structure:**

- Icon: Book icon in orange gradient circle
- Text: "Decorebator" in gradient text
- Hover: Icon rotates 12 degrees

**Logo Variations:**

```tsx
{
  /* Standard Logo */
}
;<div className="group flex cursor-pointer items-center space-x-3">
  <div className="flex h-10 w-10 transform items-center justify-center rounded-xl bg-gradient-to-br from-[#FF7B54] to-orange-600 transition-transform duration-300 group-hover:rotate-12">
    <i className="fas fa-book-open text-lg text-white"></i>
  </div>
  <span className="bg-gradient-to-r from-[#FF7B54] to-orange-600 bg-clip-text text-2xl font-bold text-transparent">
    Decorebator
  </span>
</div>
```

### Voice & Tone

**Brand Personality:**

- Modern and tech-forward
- Encouraging and supportive
- Professional yet approachable
- Educational and trustworthy

**UI Copy Guidelines:**

- Use active voice
- Keep button text action-oriented ("Start Learning Free", "Watch Demo")
- Emphasize benefits over features
- Include social proof numbers
- Use conversational tone in descriptions

### Brand Elements

**Taglines:**

- "AI-Powered Learning • 7 Languages • Multi-Platform"
- "Master Languages with [Rotating Terms]"
- Social proof: "10,000+ active learners"

**Key Messaging:**

- AI-powered vocabulary learning
- Scientifically-proven spaced repetition
- Multi-language support
- Engaging quiz modes
- Progress tracking

---

## Future Development Patterns

### Adding New Sections

**Section Template:**

```tsx
const NewSection: React.FC = () => {
  return (
    <section className="reveal bg-white py-20">
      {' '}
      {/* or bg-[#FDF6E3] for alternating */}
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="mb-16 text-center">
          <h2 className="mb-4 text-4xl font-bold lg:text-5xl">
            Section Title with
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">
              {' '}
              Highlighted Text
            </span>
          </h2>
          <p className="mx-auto max-w-3xl text-xl text-[#636E72]">Section description text</p>
        </div>

        {/* Section content */}
      </div>
    </section>
  )
}
```

### Adding New Pages

**Page Template:**

```tsx
import PageLayout from '../../components/layout/PageLayout'

export default function NewPage() {
  return (
    <PageLayout>
      <div className="pt-24 pb-20">
        <div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8">{/* Page content */}</div>
      </div>
    </PageLayout>
  )
}
```

### Animation Extensions

**Adding New Animations:**

1. Define keyframes in `src/styles/animations.css`
2. Create utility classes for common animations
3. Use consistent timing (300ms, 600ms, 3s, 6s, 8s)
4. Follow easing patterns (ease-out for entries, ease-in-out for loops)

**Staggered Animations:**

```tsx
{
  items.map((item, index) => (
    <div key={item.id} className="reveal" style={{ animationDelay: `${index * 0.1}s` }}>
      {/* Content */}
    </div>
  ))
}
```

### Component Extension Patterns

**HOC for Reveal Animations:**

```tsx
const withReveal = (Component: React.ComponentType<any>) => {
  return (props: any) => (
    <div className="reveal">
      <Component {...props} />
    </div>
  )
}
```

**Custom Hook for Animations:**

```tsx
const useScrollReveal = () => {
  useEffect(() => {
    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add('active')
        }
      })
    })

    document.querySelectorAll('.reveal').forEach((el) => {
      observer.observe(el)
    })

    return () => observer.disconnect()
  }, [])
}
```

### Performance Considerations

**Animation Performance:**

- Use `transform` and `opacity` for animations (GPU-accelerated)
- Add `will-change: transform` for frequently animated elements
- Use `animation-fill-mode: forwards` for one-time animations
- Implement `prefers-reduced-motion` media queries for accessibility

**Image Optimization:**

- Use Next.js Image component for marketing images
- Implement lazy loading for below-fold content
- Consider WebP format for better compression

**Bundle Size:**

- Keep animation CSS separate for potential lazy loading
- Use dynamic imports for large components
- Tree-shake unused Font Awesome icons

---

## Accessibility Guidelines

### Color Contrast

- Ensure 4.5:1 contrast ratio for normal text
- Test with color blindness simulators
- Don't rely solely on color for information

### Animation Accessibility

```css
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

### Keyboard Navigation

- Ensure all interactive elements are focusable
- Provide visible focus indicators
- Implement proper tab order

### Screen Reader Support

- Use semantic HTML elements
- Provide alt text for images
- Include proper ARIA labels
- Use heading hierarchy correctly

---

This design system serves as the foundation for all future development on the Decorebator landing page and related web properties. Follow these guidelines to maintain consistency and quality across all implementations.

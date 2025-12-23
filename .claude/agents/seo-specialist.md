---
name: seo-specialist
description: SEO and content optimization expert. Use when optimizing website content for search engines, implementing meta tags, setting up structured data, improving Core Web Vitals, conducting keyword research, fixing SEO issues, or auditing site performance.
---

# SEO Specialist

Optimize website content and technical infrastructure for maximum search engine visibility and ranking.

## When to Use

Activate this agent when:

- Optimizing page content for search engines
- Implementing meta tags, Open Graph, and Twitter Cards
- Setting up structured data (Schema.org, JSON-LD)
- Improving Core Web Vitals and page performance
- Conducting keyword research and analysis
- Creating SEO-friendly URLs and site structure
- Fixing SEO issues (crawl errors, duplicate content, broken links)
- Implementing canonical tags and redirects
- Optimizing images and media for SEO
- Setting up sitemaps and robots.txt
- Auditing site SEO health
- Implementing international SEO (hreflang)
- Optimizing for featured snippets and rich results

## SEO Stack

- **CMS/Framework:** Next.js, WordPress, custom HTML
- **Analytics:** Google Analytics 4, Google Search Console
- **Schema Markup:** JSON-LD, Schema.org vocabulary
- **Tools:** Lighthouse, PageSpeed Insights, Screaming Frog
- **Testing:** Mobile-Friendly Test, Rich Results Test
- **Monitoring:** Ahrefs, SEMrush, Moz (optional)
- **Performance:** Core Web Vitals, Cloudflare, CDNs

## How to Implement Meta Tags

### Essential Meta Tags

```html
<!-- Primary Meta Tags -->
<title>Page Title - Brand Name | Max 60 characters</title>
<meta name="title" content="Page Title - Brand Name" />
<meta
  name="description"
  content="Compelling description that includes target keywords. 150-160 characters max."
/>
<meta
  name="keywords"
  content="primary keyword, secondary keyword, related terms"
/>
<meta name="robots" content="index, follow" />
<meta name="language" content="English" />
<meta name="author" content="Author Name" />

<!-- Viewport for Mobile -->
<meta name="viewport" content="width=device-width, initial-scale=1.0" />

<!-- Canonical URL -->
<link rel="canonical" href="https://example.com/page" />
```

### Open Graph (Facebook, LinkedIn)

```html
<!-- Open Graph Meta Tags -->
<meta property="og:type" content="website" />
<meta property="og:url" content="https://example.com/page" />
<meta property="og:title" content="Page Title - Brand Name" />
<meta
  property="og:description"
  content="Compelling description for social sharing"
/>
<meta property="og:image" content="https://example.com/images/og-image.jpg" />
<meta property="og:image:width" content="1200" />
<meta property="og:image:height" content="630" />
<meta property="og:image:alt" content="Image description" />
<meta property="og:site_name" content="Site Name" />
<meta property="og:locale" content="en_US" />
```

### Twitter Cards

```html
<!-- Twitter Card Meta Tags -->
<meta name="twitter:card" content="summary_large_image" />
<meta name="twitter:site" content="@username" />
<meta name="twitter:creator" content="@username" />
<meta name="twitter:url" content="https://example.com/page" />
<meta name="twitter:title" content="Page Title - Brand Name" />
<meta name="twitter:description" content="Compelling description for Twitter" />
<meta
  name="twitter:image"
  content="https://example.com/images/twitter-card.jpg"
/>
<meta name="twitter:image:alt" content="Image description" />
```

### Next.js Metadata API

```typescript
// app/page.tsx or app/layout.tsx
import { Metadata } from "next";

export const metadata: Metadata = {
  title: {
    template: "%s | Brand Name",
    default: "Brand Name - Tagline",
  },
  description: "Compelling site description with target keywords",
  keywords: ["keyword1", "keyword2", "keyword3"],
  authors: [{ name: "Author Name", url: "https://example.com" }],
  creator: "Creator Name",
  publisher: "Publisher Name",
  formatDetection: {
    email: false,
    address: false,
    telephone: false,
  },
  metadataBase: new URL("https://example.com"),
  alternates: {
    canonical: "/",
    languages: {
      "en-US": "/en-US",
      "es-ES": "/es-ES",
    },
  },
  openGraph: {
    title: "Page Title",
    description: "Compelling description",
    url: "https://example.com",
    siteName: "Brand Name",
    images: [
      {
        url: "/images/og-image.jpg",
        width: 1200,
        height: 630,
        alt: "Image description",
      },
    ],
    locale: "en_US",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "Page Title",
    description: "Compelling description",
    creator: "@username",
    images: ["/images/twitter-card.jpg"],
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-video-preview": -1,
      "max-image-preview": "large",
      "max-snippet": -1,
    },
  },
  verification: {
    google: "google-site-verification-code",
    yandex: "yandex-verification-code",
    bing: "bing-verification-code",
  },
};
```

## How to Implement Structured Data

### Article Schema (Blog Posts, News)

```typescript
// app/blog/[slug]/page.tsx
export default async function BlogPost({ params }) {
  const post = await getPost(params.slug);

  const schema = {
    '@context': 'https://schema.org',
    '@type': 'Article',
    headline: post.title,
    description: post.excerpt,
    image: post.image,
    datePublished: post.publishedAt,
    dateModified: post.updatedAt,
    author: {
      '@type': 'Person',
      name: post.author.name,
      url: post.author.url,
    },
    publisher: {
      '@type': 'Organization',
      name: 'Brand Name',
      logo: {
        '@type': 'ImageObject',
        url: 'https://example.com/logo.png',
      },
    },
    mainEntityOfPage: {
      '@type': 'WebPage',
      '@id': `https://example.com/blog/${params.slug}`,
    },
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(schema) }}
      />
      <article>{/* Content */}</article>
    </>
  );
}
```

### Product Schema (E-commerce)

```typescript
const productSchema = {
  "@context": "https://schema.org",
  "@type": "Product",
  name: product.name,
  image: product.images,
  description: product.description,
  sku: product.sku,
  mpn: product.mpn,
  brand: {
    "@type": "Brand",
    name: product.brand,
  },
  offers: {
    "@type": "Offer",
    url: `https://example.com/products/${product.slug}`,
    priceCurrency: "USD",
    price: product.price,
    priceValidUntil: "2024-12-31",
    availability: "https://schema.org/InStock",
    itemCondition: "https://schema.org/NewCondition",
    seller: {
      "@type": "Organization",
      name: "Brand Name",
    },
  },
  aggregateRating: {
    "@type": "AggregateRating",
    ratingValue: product.rating,
    reviewCount: product.reviewCount,
  },
  review: product.reviews.map((review) => ({
    "@type": "Review",
    reviewRating: {
      "@type": "Rating",
      ratingValue: review.rating,
    },
    author: {
      "@type": "Person",
      name: review.author,
    },
    reviewBody: review.body,
  })),
};
```

### Organization Schema (About Page)

```typescript
const organizationSchema = {
  "@context": "https://schema.org",
  "@type": "Organization",
  name: "Brand Name",
  alternateName: "Brand Alternate Name",
  url: "https://example.com",
  logo: "https://example.com/logo.png",
  description: "Organization description",
  contactPoint: {
    "@type": "ContactPoint",
    telephone: "+1-555-555-5555",
    contactType: "customer service",
    email: "support@example.com",
    availableLanguage: ["English", "Spanish"],
  },
  sameAs: [
    "https://www.facebook.com/brandname",
    "https://twitter.com/brandname",
    "https://www.linkedin.com/company/brandname",
    "https://www.instagram.com/brandname",
  ],
  address: {
    "@type": "PostalAddress",
    streetAddress: "123 Main Street",
    addressLocality: "City",
    addressRegion: "State",
    postalCode: "12345",
    addressCountry: "US",
  },
};
```

### Breadcrumb Schema

```typescript
const breadcrumbSchema = {
  "@context": "https://schema.org",
  "@type": "BreadcrumbList",
  itemListElement: [
    {
      "@type": "ListItem",
      position: 1,
      name: "Home",
      item: "https://example.com",
    },
    {
      "@type": "ListItem",
      position: 2,
      name: "Category",
      item: "https://example.com/category",
    },
    {
      "@type": "ListItem",
      position: 3,
      name: "Product",
      item: "https://example.com/category/product",
    },
  ],
};
```

### FAQ Schema

```typescript
const faqSchema = {
  "@context": "https://schema.org",
  "@type": "FAQPage",
  mainEntity: [
    {
      "@type": "Question",
      name: "Question 1?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Detailed answer to question 1.",
      },
    },
    {
      "@type": "Question",
      name: "Question 2?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Detailed answer to question 2.",
      },
    },
  ],
};
```

## How to Optimize URLs and Site Structure

### SEO-Friendly URL Structure

```typescript
// ✅ Good URLs
/blog/how-to-optimize-seo
/products/category/product-name
/services/web-development
/about-us

// ❌ Bad URLs
/blog/post?id=123
/products/prod_xyz_123_abc
/page.php?category=1&item=456
/untitled-page-1
```

### URL Optimization Rules

1. **Use Hyphens**: Separate words with hyphens, not underscores
2. **Keep Short**: Aim for 3-5 words maximum
3. **Include Keywords**: Primary keyword near the beginning
4. **Lowercase**: Always use lowercase letters
5. **No Special Characters**: Avoid &, %, $, @, etc.
6. **Avoid Stop Words**: Remove "a", "the", "and" when possible
7. **Be Descriptive**: URL should indicate page content

### Implementing Redirects

```typescript
// next.config.js
module.exports = {
  async redirects() {
    return [
      // Permanent redirect (301)
      {
        source: "/old-page",
        destination: "/new-page",
        permanent: true,
      },
      // Temporary redirect (302)
      {
        source: "/temporary",
        destination: "/temp-destination",
        permanent: false,
      },
      // Wildcard redirect
      {
        source: "/blog/:slug*",
        destination: "/news/:slug*",
        permanent: true,
      },
      // Regex redirect
      {
        source: "/old-blog/:year(\\d{4})/:month(\\d{2})/:slug",
        destination: "/blog/:slug",
        permanent: true,
      },
    ];
  },
};
```

## How to Create Sitemaps

### XML Sitemap (Next.js)

```typescript
// app/sitemap.ts
import { MetadataRoute } from "next";

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const baseUrl = "https://example.com";

  // Static pages
  const staticPages = [
    {
      url: baseUrl,
      lastModified: new Date(),
      changeFrequency: "daily" as const,
      priority: 1,
    },
    {
      url: `${baseUrl}/about`,
      lastModified: new Date(),
      changeFrequency: "monthly" as const,
      priority: 0.8,
    },
  ];

  // Dynamic pages (blog posts)
  const posts = await getAllPosts();
  const postPages = posts.map((post) => ({
    url: `${baseUrl}/blog/${post.slug}`,
    lastModified: new Date(post.updatedAt),
    changeFrequency: "weekly" as const,
    priority: 0.7,
  }));

  // Dynamic pages (products)
  const products = await getAllProducts();
  const productPages = products.map((product) => ({
    url: `${baseUrl}/products/${product.slug}`,
    lastModified: new Date(product.updatedAt),
    changeFrequency: "daily" as const,
    priority: 0.9,
  }));

  return [...staticPages, ...postPages, ...productPages];
}
```

### Robots.txt

```typescript
// app/robots.ts
import { MetadataRoute } from "next";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: "*",
        allow: "/",
        disallow: ["/admin/", "/api/", "/private/"],
      },
      {
        userAgent: "Googlebot",
        allow: "/",
        disallow: ["/admin/", "/private/"],
      },
    ],
    sitemap: "https://example.com/sitemap.xml",
    host: "https://example.com",
  };
}
```

## How to Optimize Images for SEO

### Image Optimization Checklist

```typescript
import Image from 'next/image';

// ✅ SEO-Optimized Image
<Image
  src="/images/product-name-feature.jpg"  // Descriptive filename
  alt="Product Name showing key feature in detail"  // Descriptive alt text
  width={800}
  height={600}
  quality={85}  // Balance quality vs file size
  priority={false}  // Only true for above-the-fold
  loading="lazy"  // Lazy load by default
  sizes="(max-width: 768px) 100vw, 800px"  // Responsive sizing
  title="Product Name Feature"  // Optional title
/>

// Image Schema
const imageSchema = {
  '@context': 'https://schema.org',
  '@type': 'ImageObject',
  contentUrl: 'https://example.com/images/product.jpg',
  description: 'Detailed image description',
  name: 'Image title',
  author: {
    '@type': 'Organization',
    name: 'Brand Name',
  },
  copyrightNotice: 'Copyright 2024 Brand Name',
  license: 'https://example.com/license',
};
```

### Image Best Practices

1. **Format**: Use WebP with JPEG/PNG fallback
2. **Size**: Keep under 200KB, compress without quality loss
3. **Dimensions**: Serve responsive sizes based on viewport
4. **Alt Text**: Descriptive, include keywords naturally
5. **Filename**: Use hyphens, be descriptive (product-blue-widget.jpg)
6. **Lazy Loading**: Enable for below-the-fold images
7. **Compression**: 80-85% quality for photos, lossless for graphics
8. **CDN**: Serve from CDN for faster loading

## How to Optimize Core Web Vitals

### Largest Contentful Paint (LCP) - Target: <2.5s

```typescript
// Optimize with priority loading
<Image
  src="/hero-image.jpg"
  priority  // Preload above-the-fold images
  fetchPriority="high"
  alt="Hero image"
/>

// Preload critical resources
<link rel="preload" href="/fonts/font.woff2" as="font" type="font/woff2" crossOrigin="anonymous" />

// Use Server Components for data fetching
async function HeroSection() {
  const data = await fetch('https://api.example.com/hero', {
    next: { revalidate: 3600 }  // Cache for 1 hour
  });

  return <Hero data={data} />;
}
```

### First Input Delay (FID) - Target: <100ms

```typescript
// Code splitting with dynamic imports
import dynamic from 'next/dynamic';

const HeavyComponent = dynamic(() => import('@/components/HeavyComponent'), {
  loading: () => <Skeleton />,
  ssr: false,  // Skip SSR for client-only components
});

// Defer non-critical JavaScript
<script src="/analytics.js" defer />

// Use web workers for heavy computations
const worker = new Worker('/heavy-computation.worker.js');
```

### Cumulative Layout Shift (CLS) - Target: <0.1

```typescript
// Always specify image dimensions
<Image src="/image.jpg" width={800} height={600} alt="Description" />

// Reserve space for dynamic content
<div className="min-h-[200px]">
  <Suspense fallback={<div className="h-[200px] bg-gray-200" />}>
    <DynamicContent />
  </Suspense>
</div>

// Use aspect-ratio for responsive elements
<div className="aspect-[16/9]">
  <iframe src="..." className="w-full h-full" />
</div>

// Avoid inserting content above existing content
// Use transform instead of top/left for animations
<div className="transition-transform hover:translate-y-2">
  Content
</div>
```

## How to Conduct Keyword Research

### Keyword Research Process

1. **Identify Seed Keywords**
   - Primary topics and products
   - Industry terms and jargon
   - Customer language from support tickets

2. **Expand with Tools**
   - Google Keyword Planner
   - Google Search Console
   - Answer The Public
   - Competitor analysis

3. **Analyze Metrics**
   - Search volume (monthly searches)
   - Keyword difficulty (competition)
   - Search intent (informational, navigational, transactional)
   - CPC (if running ads)

4. **Categorize Keywords**
   - **Head terms**: High volume, high competition (1-2 words)
   - **Body terms**: Medium volume, medium competition (2-3 words)
   - **Long-tail**: Low volume, low competition (4+ words)

5. **Match to Content**
   - Homepage: Brand + primary service
   - Category pages: Body terms
   - Product/blog pages: Long-tail keywords

### Keyword Placement Strategy

```html
<!-- Title Tag: Primary keyword at beginning -->
<title>Primary Keyword - Secondary Keyword | Brand</title>

<!-- H1: Include primary keyword -->
<h1>Complete Guide to Primary Keyword in 2024</h1>

<!-- First paragraph: Use keyword within first 100 words -->
<p>When it comes to primary keyword, understanding the fundamentals...</p>

<!-- Subheadings: Use variations and related keywords -->
<h2>Why Primary Keyword Matters for Your Business</h2>
<h3>Related Keyword Best Practices</h3>

<!-- Body content: Natural keyword density 1-2% -->
<!-- Use LSI keywords and synonyms throughout -->

<!-- Image alt text: Include keywords naturally -->
<img
  src="primary-keyword-guide.jpg"
  alt="Primary keyword implementation guide"
/>

<!-- Internal links: Use keyword-rich anchor text -->
<a href="/related-page">Learn more about related keyword</a>

<!-- URL: Include primary keyword -->
https://example.com/primary-keyword-guide
```

## How to Implement International SEO

### Hreflang Tags

```html
<!-- For pages with multiple language versions -->
<link rel="alternate" hreflang="en" href="https://example.com/en/page" />
<link rel="alternate" hreflang="es" href="https://example.com/es/page" />
<link rel="alternate" hreflang="fr" href="https://example.com/fr/page" />
<link rel="alternate" hreflang="x-default" href="https://example.com/page" />

<!-- For regional variations -->
<link rel="alternate" hreflang="en-US" href="https://example.com/us/page" />
<link rel="alternate" hreflang="en-GB" href="https://example.com/uk/page" />
<link rel="alternate" hreflang="en-AU" href="https://example.com/au/page" />
```

### Next.js Implementation

```typescript
// app/[lang]/layout.tsx
export async function generateMetadata({ params }) {
  const languages = ["en", "es", "fr", "de"];

  return {
    alternates: {
      canonical: `https://example.com/${params.lang}`,
      languages: Object.fromEntries(
        languages.map((lang) => [lang, `https://example.com/${lang}`]),
      ),
    },
  };
}
```

## How to Fix Common SEO Issues

### Duplicate Content

```typescript
// Set canonical URL to preferred version
export const metadata = {
  alternates: {
    canonical: 'https://example.com/preferred-url',
  },
};

// Consolidate similar pages
// Use 301 redirects from duplicates to canonical

// For pagination, use rel="next" and rel="prev"
<link rel="prev" href="https://example.com/page/1" />
<link rel="next" href="https://example.com/page/3" />

// For parameter-based URLs, use canonical
// https://example.com/products?sort=price (canonical: /products)
```

### Broken Links (404s)

```typescript
// Implement custom 404 page with helpful navigation
// app/not-found.tsx
export default function NotFound() {
  return (
    <div>
      <h1>404 - Page Not Found</h1>
      <p>The page you're looking for doesn't exist.</p>
      <nav>
        <a href="/">Home</a>
        <a href="/products">Products</a>
        <a href="/contact">Contact Us</a>
      </nav>

      {/* Log 404s for monitoring */}
      <script dangerouslySetInnerHTML={{
        __html: `
          if (typeof window !== 'undefined') {
            fetch('/api/log-404', {
              method: 'POST',
              body: JSON.stringify({ url: window.location.href })
            });
          }
        `
      }} />
    </div>
  );
}

// Set up 301 redirects in next.config.js for known moved pages
```

### Slow Page Speed

```typescript
// 1. Optimize images
<Image src="..." quality={85} loading="lazy" />

// 2. Enable compression
// next.config.js
module.exports = {
  compress: true,
  images: {
    formats: ['image/avif', 'image/webp'],
  },
};

// 3. Minimize JavaScript
// Use dynamic imports for code splitting
const Component = dynamic(() => import('./Component'));

// 4. Optimize fonts
import { Inter } from 'next/font/google';
const inter = Inter({ subsets: ['latin'], display: 'swap' });

// 5. Use CDN for static assets
// 6. Enable HTTP/2 on server
// 7. Implement caching headers
```

## Quality Checklist

Before launching SEO-optimized pages:

- [ ] Title tag includes primary keyword (50-60 characters)
- [ ] Meta description is compelling (150-160 characters)
- [ ] H1 tag includes primary keyword (only one H1 per page)
- [ ] URL is SEO-friendly and includes keyword
- [ ] Images have descriptive alt text
- [ ] Images are optimized (WebP, compressed, responsive)
- [ ] Structured data (JSON-LD) is implemented and validated
- [ ] Canonical tag is set correctly
- [ ] Internal links use descriptive anchor text
- [ ] Page loads in under 3 seconds
- [ ] Core Web Vitals are in "Good" range
- [ ] Mobile-friendly (responsive design)
- [ ] HTTPS is enabled
- [ ] No broken links (404s)
- [ ] No duplicate content
- [ ] XML sitemap includes page
- [ ] Robots.txt allows crawling
- [ ] Open Graph and Twitter Cards configured
- [ ] Hreflang tags for international versions (if applicable)
- [ ] Content is unique and valuable (500+ words)
- [ ] Keyword density is natural (1-2%)

## Tools and Validation

### Testing Tools

```bash
# Google Rich Results Test
https://search.google.com/test/rich-results

# Mobile-Friendly Test
https://search.google.com/test/mobile-friendly

# PageSpeed Insights
https://pagespeed.web.dev/

# Lighthouse (in Chrome DevTools)
# Run: DevTools → Lighthouse → Generate report

# Schema Markup Validator
https://validator.schema.org/

# Screaming Frog (Desktop app for crawling)
# Check: Dead links, duplicate content, redirects
```

### Monitoring

```typescript
// Google Search Console integration
// Add verification meta tag
<meta name="google-site-verification" content="verification-code" />

// Track in Google Analytics 4
<script async src="https://www.googletagmanager.com/gtag/js?id=G-XXXXXXXXXX"></script>

// Monitor Core Web Vitals with web-vitals library
import { getCLS, getFID, getFCP, getLCP, getTTFB } from 'web-vitals';

function sendToAnalytics(metric) {
  const body = JSON.stringify(metric);
  // Send to analytics endpoint
  fetch('/api/analytics', { method: 'POST', body });
}

getCLS(sendToAnalytics);
getFID(sendToAnalytics);
getFCP(sendToAnalytics);
getLCP(sendToAnalytics);
getTTFB(sendToAnalytics);
```

## Example Workflows

### Optimizing a New Blog Post

1. Conduct keyword research for topic
2. Choose primary keyword (medium volume, low competition)
3. Create SEO-friendly URL with keyword
4. Write compelling title (keyword at beginning, <60 chars)
5. Write meta description (keyword included, <160 chars)
6. Structure content with H1 (primary keyword) and H2/H3 (variations)
7. Write 1000+ words of original, valuable content
8. Optimize images (compress, descriptive filenames, alt text)
9. Add internal links to related content
10. Implement Article schema with JSON-LD
11. Add Open Graph and Twitter Card meta tags
12. Validate with Rich Results Test
13. Submit URL to Google Search Console
14. Monitor rankings and traffic

### Conducting an SEO Audit

1. Crawl site with Screaming Frog or similar tool
2. Identify technical issues (404s, redirects, duplicates)
3. Check robots.txt and sitemap.xml
4. Run Lighthouse audit on key pages
5. Analyze Core Web Vitals in Search Console
6. Review top pages in Google Analytics
7. Check mobile usability in Search Console
8. Validate structured data with Rich Results Test
9. Review internal linking structure
10. Analyze top keywords and rankings
11. Check backlink profile quality
12. Create prioritized action plan
13. Fix critical issues first (broken links, slow pages)
14. Optimize underperforming pages
15. Monitor improvements over 30-90 days

### Launching a New Product Page

1. Research product-specific keywords
2. Optimize product title and description
3. Create SEO-friendly URL
4. Implement Product schema with price, availability, reviews
5. Optimize product images (high quality, multiple angles)
6. Write unique product description (avoid manufacturer copy)
7. Add FAQ section with FAQ schema
8. Include related products (internal linking)
9. Add breadcrumb navigation with schema
10. Implement review/rating system
11. Test page speed and Core Web Vitals
12. Validate all structured data
13. Submit to Search Console
14. Monitor impressions and clicks

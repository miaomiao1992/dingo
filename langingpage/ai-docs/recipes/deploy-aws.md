# Deploy Astro to AWS

## Overview

AWS (Amazon Web Services) provides multiple ways to deploy Astro sites, from fully-managed platforms to manual infrastructure setup. This guide covers three main deployment methods: AWS Amplify (easiest), S3 Static Hosting (flexible), and CloudFront with S3 (production-ready).

## Deployment Options

### AWS Amplify

**Best for:**
- Quick deployment with CI/CD
- Teams familiar with AWS services
- Both static and SSR sites

**Pros:**
- Automatic deployments from Git
- Built-in CI/CD pipeline
- Free tier available
- Easy setup

**Cons:**
- AWS-specific platform
- Less control than S3

### S3 Static Hosting

**Best for:**
- Simple static sites
- Budget-conscious projects
- Learning AWS basics

**Pros:**
- Very low cost
- Simple setup
- Reliable hosting

**Cons:**
- Manual deployment process
- No HTTPS by default
- No CDN without CloudFront

### CloudFront + S3

**Best for:**
- Production sites
- Global audiences
- Performance-critical applications

**Pros:**
- Global CDN distribution
- HTTPS support
- Great performance
- Cost-effective at scale

**Cons:**
- More complex setup
- Cache invalidation needed

## AWS Amplify Deployment

### Static Site (Default)

**1. Build Your Site:**

```bash
pnpm build
```

This creates a `dist/` folder with your static files.

**2. Create Amplify Project:**

- Go to [AWS Amplify Console](https://console.aws.amazon.com/amplify)
- Click "New app" → "Host web app"
- Choose your Git provider (GitHub, GitLab, Bitbucket, CodeCommit)
- Authorize AWS to access your repository
- Select the repository and branch

**3. Configure Build Settings:**

Amplify will auto-detect your framework. Verify or update the build settings:

**For npm:**

```yaml
version: 1
frontend:
  phases:
    preBuild:
      commands:
        - npm ci
    build:
      commands:
        - npm run build
  artifacts:
    baseDirectory: /dist
    files:
      - '**/*'
  cache:
    paths:
      - node_modules/**/*
```

**For pnpm:**

```yaml
version: 1
frontend:
  phases:
    preBuild:
      commands:
        - npm i -g pnpm
        - pnpm config set store-dir .pnpm-store
        - pnpm i
    build:
      commands:
        - pnpm run build
  artifacts:
    baseDirectory: /dist
    files:
      - '**/*'
  cache:
    paths:
      - .pnpm-store/**/*
```

**For Yarn:**

```yaml
version: 1
frontend:
  phases:
    preBuild:
      commands:
        - yarn install
    build:
      commands:
        - yarn build
  artifacts:
    baseDirectory: /dist
    files:
      - '**/*'
  cache:
    paths:
      - node_modules/**/*
```

**4. Environment Variables (if needed):**

- Click "Environment variables" tab
- Add any required variables (e.g., `PUBLIC_API_URL`)

**5. Deploy:**

- Click "Save and deploy"
- Amplify will automatically build and deploy your site
- Future commits to your branch will trigger automatic deployments

### SSR Site with Amplify Adapter

For server-side rendering (SSR) with Astro:

**1. Install Adapter:**

```bash
pnpm add astro-aws-amplify
```

**2. Update Config:**

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import awsAmplify from 'astro-aws-amplify';

export default defineConfig({
  output: 'server',
  adapter: awsAmplify(),
});
```

**3. Create Amplify Project:**

Follow the same steps as static site deployment.

**4. Update Build Settings:**

**For npm:**

```yaml
version: 1
frontend:
  phases:
    preBuild:
      commands:
        - npm ci
    build:
      commands:
        - npm run build
        - mv node_modules ./.amplify-hosting/compute/default
  artifacts:
    baseDirectory: .amplify-hosting
    files:
      - '**/*'
  cache:
    paths:
      - node_modules/**/*
```

**For pnpm:**

```yaml
version: 1
frontend:
  phases:
    preBuild:
      commands:
        - npm i -g pnpm
        - pnpm config set store-dir .pnpm-store
        - pnpm i
    build:
      commands:
        - pnpm run build
        - mv node_modules ./.amplify-hosting/compute/default
  artifacts:
    baseDirectory: .amplify-hosting
    files:
      - '**/*'
  cache:
    paths:
      - .pnpm-store/**/*
```

**5. Deploy:**

Commit and push your changes. Amplify will automatically deploy your SSR site.

## S3 Static Hosting

### Manual Setup

**1. Create S3 Bucket:**

```bash
# Replace <BUCKET_NAME> with your desired name (must be globally unique)
aws s3 mb s3://<BUCKET_NAME>
```

Or via AWS Console:
- Go to S3 → Create bucket
- Choose a globally unique name (e.g., `my-astro-site-2025`)
- Select a region close to your users
- **Uncheck** "Block all public access"
- Create bucket

**2. Build Your Site:**

```bash
pnpm build
```

**3. Upload Files:**

**Using AWS CLI:**

```bash
aws s3 sync dist/ s3://<BUCKET_NAME>/ --delete
```

**Using AWS Console:**
- Navigate to your bucket
- Click "Upload"
- Drag and drop files from `dist/` folder

**4. Configure Bucket Policy:**

- Go to bucket → Permissions → Bucket policy
- Add the following policy:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicReadGetObject",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::<BUCKET_NAME>/*"
    }
  ]
}
```

⚠️ **Replace `<BUCKET_NAME>` with your actual bucket name.**

**5. Enable Static Website Hosting:**

- Go to bucket → Properties → Static website hosting
- Click "Edit"
- Enable static website hosting
- Index document: `index.html`
- Error document: `404.html` (or `index.html` for SPAs)
- Save changes

**6. Access Your Site:**

Find your website endpoint in Properties → Static website hosting → Bucket website endpoint.

Example: `http://my-astro-site-2025.s3-website-us-east-1.amazonaws.com`

### Limitations

- HTTP only (no HTTPS)
- No custom domain without Route 53
- No CDN (slower for global users)

## CloudFront + S3 (Production Setup)

### Setup CloudFront Distribution

**1. Complete S3 Static Hosting Setup:**

Follow all steps in the S3 Static Hosting section above.

**2. Create CloudFront Distribution:**

- Go to CloudFront → Create distribution
- **Origin domain**: Click the field and select your S3 bucket

  ⚠️ **Important:** Use the static website endpoint, not the S3 bucket directly:
  - Good: `my-bucket.s3-website-us-east-1.amazonaws.com`
  - Bad: `my-bucket.s3.amazonaws.com`

  Alternatively, paste your S3 static website endpoint from Properties → Static website hosting.

- **Origin path**: Leave empty
- **Viewer protocol policy**: "Redirect HTTP to HTTPS"
- **Allowed HTTP methods**: GET, HEAD
- **Cache policy**: CachingOptimized (recommended)
- **WAF**: Do not enable (unless needed)

**3. Configure Error Pages (for SPAs):**

- Go to distribution → Error pages → Create custom error response
- HTTP error code: 404
- Customize error response: Yes
- Response page path: `/index.html`
- HTTP response code: 200
- Create error response

**4. Wait for Deployment:**

- Status will change from "In Progress" to "Enabled"
- Takes 5-15 minutes
- Find your CloudFront URL in the "Distribution domain name" field
  - Example: `d1234abcd.cloudfront.net`

**5. Test Your Site:**

Visit `https://<distribution-domain-name>` to verify deployment.

### Custom Domain with CloudFront

**1. Request SSL Certificate (ACM):**

⚠️ **Must be in us-east-1 region for CloudFront.**

- Go to AWS Certificate Manager → Request certificate
- Request a public certificate
- Add domain names:
  - `example.com`
  - `*.example.com` (for subdomains)
- Validation method: DNS validation (recommended)
- Request certificate

**2. Validate Domain:**

- Click "Create records in Route 53" (if using Route 53)
- Or manually add CNAME records to your DNS provider
- Wait for validation (can take minutes to hours)

**3. Update CloudFront Distribution:**

- Go to distribution → Settings → Edit
- **Alternate domain names (CNAMEs)**: Add your domain (e.g., `example.com`)
- **Custom SSL certificate**: Select your ACM certificate
- Save changes

**4. Update DNS:**

Add a CNAME or ALIAS record pointing to your CloudFront distribution:

**Route 53:**
- Create ALIAS record
- Type: A - IPv4 address
- Route traffic to: Alias to CloudFront distribution
- Select your distribution

**Other DNS Providers:**
- Create CNAME record
- Name: `@` or `www`
- Value: Your CloudFront domain name
- TTL: 300

**5. Update Astro Config:**

```javascript
// astro.config.mjs
export default defineConfig({
  site: 'https://example.com',
});
```

## CI/CD with GitHub Actions

### Setup

**1. Create IAM Policy:**

- Go to IAM → Policies → Create policy
- Use JSON editor:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "S3AndCloudFrontAccess",
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:ListBucket",
        "s3:DeleteObject",
        "cloudfront:CreateInvalidation"
      ],
      "Resource": [
        "arn:aws:cloudfront::<AWS_ACCOUNT_ID>:distribution/<DISTRIBUTION_ID>",
        "arn:aws:s3:::<BUCKET_NAME>/*",
        "arn:aws:s3:::<BUCKET_NAME>"
      ]
    }
  ]
}
```

⚠️ **Replace:**
- `<AWS_ACCOUNT_ID>`: Your AWS account ID
- `<DISTRIBUTION_ID>`: Your CloudFront distribution ID
- `<BUCKET_NAME>`: Your S3 bucket name

**2. Create IAM User:**

- Go to IAM → Users → Create user
- User name: `astro-deployer`
- Attach the policy created above
- Click "Create user"

**3. Create Access Keys:**

- Go to user → Security credentials → Create access key
- Use case: Application running outside AWS
- Create access key
- **Save** `Access key ID` and `Secret access key`

**4. Add GitHub Secrets:**

- Go to your repository → Settings → Secrets and variables → Actions
- Add the following secrets:
  - `AWS_ACCESS_KEY_ID`: Your access key ID
  - `AWS_SECRET_ACCESS_KEY`: Your secret access key
  - `BUCKET_ID`: Your S3 bucket name
  - `DISTRIBUTION_ID`: Your CloudFront distribution ID

**5. Create Workflow File:**

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy to AWS

on:
  push:
    branches:
      - main

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: '22'

      - name: Setup pnpm
        uses: pnpm/action-setup@v3
        with:
          version: 9

      - name: Install dependencies
        run: pnpm install

      - name: Build
        run: pnpm run build

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

      - name: Deploy to S3
        run: aws s3 sync --delete ./dist/ s3://${{ secrets.BUCKET_ID }}

      - name: Invalidate CloudFront Cache
        run: |
          aws cloudfront create-invalidation \
            --distribution-id ${{ secrets.DISTRIBUTION_ID }} \
            --paths "/*"
```

**6. Deploy:**

- Commit and push the workflow file
- GitHub Actions will automatically deploy on every push to `main`

## Cost Estimation

### AWS Amplify

**Free Tier:**
- Build minutes: 1,000 per month
- Hosting: 5 GB stored, 15 GB served per month

**After Free Tier:**
- Build minutes: $0.01/minute
- Hosting: $0.15/GB served

**Example:** Small blog (~100MB, 10,000 monthly visitors):
- Likely free under free tier
- After: ~$2-5/month

### S3 + CloudFront

**S3:**
- Storage: $0.023/GB/month
- PUT requests: $0.005/1,000 requests
- GET requests: $0.0004/1,000 requests

**CloudFront:**
- Data transfer: $0.085/GB (first 10 TB/month)
- Requests: $0.0075/10,000 HTTPS requests

**Example:** Small blog (~100MB, 10,000 monthly visitors):
- S3: ~$0.10/month
- CloudFront: ~$1-2/month
- **Total: ~$1-3/month**

## Best Practices

### 1. Use Environment Variables

```yaml
# amplify.yml or .github/workflows/deploy.yml
env:
  PUBLIC_API_URL: ${{ secrets.PUBLIC_API_URL }}
  DATABASE_URL: ${{ secrets.DATABASE_URL }}
```

### 2. Cache Invalidation

Always invalidate CloudFront cache after deployment:

```bash
aws cloudfront create-invalidation \
  --distribution-id <DISTRIBUTION_ID> \
  --paths "/*"
```

### 3. Compression

CloudFront automatically compresses files. Ensure your `astro.config.mjs` enables compression:

```javascript
export default defineConfig({
  vite: {
    build: {
      minify: 'esbuild',
      cssMinify: true,
    },
  },
});
```

### 4. Security Headers

Add security headers via CloudFront Functions or Lambda@Edge:

```javascript
// CloudFront Function
function handler(event) {
  var response = event.response;
  response.headers['strict-transport-security'] = {
    value: 'max-age=63072000'
  };
  response.headers['x-content-type-options'] = { value: 'nosniff' };
  response.headers['x-frame-options'] = { value: 'DENY' };
  return response;
}
```

### 5. Monitoring

Enable CloudFront logging:
- Go to distribution → Settings
- Enable standard logging
- Choose S3 bucket for logs

## Troubleshooting

### Issue: 403 Forbidden

**Cause:** Bucket policy not set or incorrect.

**Solution:**
1. Verify bucket policy allows public access
2. Check "Block public access" settings are disabled

### Issue: CloudFront Serves Old Content

**Cause:** Cache not invalidated.

**Solution:**
```bash
aws cloudfront create-invalidation \
  --distribution-id <DISTRIBUTION_ID> \
  --paths "/*"
```

### Issue: 404 on SPA Routes

**Cause:** S3 or CloudFront not configured for SPA routing.

**Solution:**
- S3: Set error document to `index.html`
- CloudFront: Create custom error response mapping 404 to `/index.html` with 200 status

### Issue: Build Fails in Amplify

**Cause:** Missing dependencies or incorrect build command.

**Solution:**
1. Check build logs in Amplify console
2. Verify `amplify.yml` build commands
3. Ensure all dependencies in `package.json`

## Quick Reference

### Deploy to Amplify (Static)

```bash
# 1. Push to Git
git push origin main

# 2. Amplify automatically builds and deploys
```

### Deploy to S3

```bash
# 1. Build
pnpm build

# 2. Upload
aws s3 sync dist/ s3://<BUCKET_NAME>/ --delete
```

### Deploy to CloudFront

```bash
# 1. Build
pnpm build

# 2. Upload to S3
aws s3 sync dist/ s3://<BUCKET_NAME>/ --delete

# 3. Invalidate cache
aws cloudfront create-invalidation \
  --distribution-id <DISTRIBUTION_ID> \
  --paths "/*"
```

## See Also

- [04-development-workflow.md](../04-development-workflow.md) - Build process
- [recipes/deploy-github-pages.md](./deploy-github-pages.md) - GitHub Pages alternative
- [AWS Amplify Docs](https://docs.amplify.aws/)
- [AWS S3 Docs](https://docs.aws.amazon.com/s3/)
- [CloudFront Docs](https://docs.aws.amazon.com/cloudfront/)

---

**Last Updated**: 2025-11-17
**Purpose**: Comprehensive guide to deploying Astro sites on AWS
**Key Concepts**: AWS Amplify, S3 static hosting, CloudFront CDN, CI/CD with GitHub Actions, custom domains

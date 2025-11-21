// Firebase Authentication Configuration
// Reference: ai-docs/02-islands-architecture.md - Client-side Firebase SDK

import { initializeApp } from 'firebase/app';
import {
  getAuth,
  signInWithPopup,
  signOut as firebaseSignOut,
  GoogleAuthProvider,
  GithubAuthProvider,
  type Auth,
  type UserCredential
} from 'firebase/auth';
import { getFirestore, type Firestore } from 'firebase/firestore';

// Firebase configuration from environment variables
// Uses PUBLIC_ prefix to expose to client-side (Astro requirement)
const firebaseConfig = {
  apiKey: import.meta.env.PUBLIC_FIREBASE_API_KEY,
  authDomain: import.meta.env.PUBLIC_FIREBASE_AUTH_DOMAIN,
  projectId: import.meta.env.PUBLIC_FIREBASE_PROJECT_ID,
  storageBucket: import.meta.env.PUBLIC_FIREBASE_STORAGE_BUCKET,
  messagingSenderId: import.meta.env.PUBLIC_FIREBASE_MESSAGING_SENDER_ID,
  appId: import.meta.env.PUBLIC_FIREBASE_APP_ID,
};

// Initialize Firebase (client-side only)
// Note: Initialization is safe here because this module only imports in client-side components
let app: ReturnType<typeof initializeApp> | undefined;
let _auth: Auth | undefined;
let _db: Firestore | undefined;

if (typeof window !== 'undefined') {
  app = initializeApp(firebaseConfig);
  _auth = getAuth(app);
  _db = getFirestore(app);
}

// Export initialized instances
// These will be undefined during SSR (which is fine - React components guard with client: directives)
export const auth = _auth as Auth;
export const db = _db as Firestore;

// Auth Providers (client-side only)
let googleProvider: GoogleAuthProvider | undefined;
let githubProvider: GithubAuthProvider | undefined;

if (typeof window !== 'undefined') {
  googleProvider = new GoogleAuthProvider();
  githubProvider = new GithubAuthProvider();
}

// Sign in with Google
export async function signInWithGoogle(): Promise<UserCredential> {
  if (!googleProvider) {
    throw new Error('Firebase not initialized');
  }
  return signInWithPopup(auth, googleProvider);
}

// Sign in with GitHub
export async function signInWithGitHub(): Promise<UserCredential> {
  if (!githubProvider) {
    throw new Error('Firebase not initialized');
  }
  return signInWithPopup(auth, githubProvider);
}

// Sign out
export async function signOut(): Promise<void> {
  return firebaseSignOut(auth);
}

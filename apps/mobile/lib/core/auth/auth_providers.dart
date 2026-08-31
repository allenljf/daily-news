import 'package:firebase_auth/firebase_auth.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_sign_in/google_sign_in.dart';

import 'auth_gateway.dart';
import 'auth_repository.dart';
import 'auth_user.dart';
import 'firebase_google_auth_gateway.dart';

final firebaseAuthProvider = Provider<FirebaseAuth>(
  (ref) => FirebaseAuth.instance,
);

final googleSignInProvider = Provider<GoogleSignIn>(
  (ref) => GoogleSignIn.instance,
);

final googleSignInInitializationProvider = Provider<Future<void>>(
  (ref) => ref.watch(googleSignInProvider).initialize(),
);

final authGatewayProvider = Provider<AuthGateway>(
  (ref) => FirebaseGoogleAuthGateway(
    firebaseAuth: ref.watch(firebaseAuthProvider),
    googleSignIn: ref.watch(googleSignInProvider),
    googleSignInInitialized: ref.watch(googleSignInInitializationProvider),
  ),
);

final authRepositoryProvider = Provider<AuthRepository>(
  (ref) => AuthRepository(ref.watch(authGatewayProvider)),
);

final authStateProvider = StreamProvider<AuthUser?>(
  (ref) => ref.watch(authRepositoryProvider).authStateChanges(),
);

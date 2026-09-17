import 'package:firebase_core/firebase_core.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_sign_in/google_sign_in.dart';

import 'auth_gateway.dart';
import 'auth_repository.dart';
import 'auth_user.dart';
import 'firebase_google_auth_gateway.dart';

// Resolves after Firebase.initializeApp(), which runs off the startup critical
// path. Anything that needs FirebaseAuth must await this first.
final firebaseInitializationProvider = FutureProvider<void>(
  (ref) async {
    await Firebase.initializeApp();
  },
);

final googleSignInProvider = Provider<GoogleSignIn>(
  (ref) => GoogleSignIn.instance,
);

final googleSignInInitializationProvider = Provider<Future<void>>(
  (ref) => ref.watch(googleSignInProvider).initialize(),
);

final authGatewayProvider = Provider<AuthGateway>(
  (ref) => FirebaseGoogleAuthGateway(
    firebaseReady: ref.watch(firebaseInitializationProvider.future),
    googleSignIn: ref.watch(googleSignInProvider),
    // Read on demand so startup and tests never touch the Google plugin until
    // the user actually signs in or out.
    googleSignInInitializer: () =>
        ref.read(googleSignInInitializationProvider),
  ),
);

final authRepositoryProvider = Provider<AuthRepository>(
  (ref) => AuthRepository(ref.watch(authGatewayProvider)),
);

final authStateProvider = StreamProvider<AuthUser?>(
  (ref) => ref.watch(authRepositoryProvider).authStateChanges(),
);

import 'package:firebase_auth/firebase_auth.dart';
import 'package:google_sign_in/google_sign_in.dart';

import 'auth_gateway.dart';
import 'auth_user.dart';

final class FirebaseGoogleAuthGateway implements AuthGateway {
  FirebaseGoogleAuthGateway({
    required Future<void> firebaseReady,
    required GoogleSignIn googleSignIn,
    required Future<void> Function() googleSignInInitializer,
  }) : this._(firebaseReady, googleSignIn, googleSignInInitializer);

  FirebaseGoogleAuthGateway._(
    this._firebaseReady,
    this._googleSignIn,
    this._googleSignInInitializer,
  );

  final Future<void> _firebaseReady;
  final GoogleSignIn _googleSignIn;
  final Future<void> Function() _googleSignInInitializer;

  // FirebaseAuth.instance is only safe to read once Firebase.initializeApp()
  // has completed, which happens off the startup critical path.
  Future<FirebaseAuth> get _firebaseAuth async {
    await _firebaseReady;
    return FirebaseAuth.instance;
  }

  @override
  Stream<AuthUser?> authStateChanges() async* {
    final firebaseAuth = await _firebaseAuth;
    yield* firebaseAuth.authStateChanges().map(
      (user) => user == null ? null : AuthUser(id: user.uid),
    );
  }

  @override
  Future<String?> getIdToken() async =>
      (await _firebaseAuth).currentUser?.getIdToken();

  @override
  Future<void> signInWithGoogle() async {
    await _googleSignInInitializer();
    final googleAccount = await _googleSignIn.authenticate();
    final idToken = googleAccount.authentication.idToken;
    if (idToken == null) {
      throw const AuthConfigurationException();
    }

    final credential = GoogleAuthProvider.credential(idToken: idToken);
    await (await _firebaseAuth).signInWithCredential(credential);
  }

  @override
  Future<void> signOut() async {
    await (await _firebaseAuth).signOut();
    await _googleSignInInitializer();
    await _googleSignIn.signOut();
  }
}

final class AuthConfigurationException implements Exception {
  const AuthConfigurationException();
}

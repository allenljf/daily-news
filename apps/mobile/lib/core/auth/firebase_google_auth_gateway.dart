import 'package:firebase_auth/firebase_auth.dart';
import 'package:google_sign_in/google_sign_in.dart';

import 'auth_gateway.dart';
import 'auth_user.dart';

final class FirebaseGoogleAuthGateway implements AuthGateway {
  FirebaseGoogleAuthGateway({
    required FirebaseAuth firebaseAuth,
    required GoogleSignIn googleSignIn,
    required Future<void> googleSignInInitialized,
  }) : this._(firebaseAuth, googleSignIn, googleSignInInitialized);

  FirebaseGoogleAuthGateway._(
    this._firebaseAuth,
    this._googleSignIn,
    this._googleSignInInitialized,
  );

  final FirebaseAuth _firebaseAuth;
  final GoogleSignIn _googleSignIn;
  final Future<void> _googleSignInInitialized;

  @override
  Stream<AuthUser?> authStateChanges() => _firebaseAuth.authStateChanges().map(
    (user) => user == null ? null : AuthUser(id: user.uid),
  );

  @override
  Future<String?> getIdToken() async => _firebaseAuth.currentUser?.getIdToken();

  @override
  Future<void> signInWithGoogle() async {
    await _googleSignInInitialized;
    final googleAccount = await _googleSignIn.authenticate();
    final idToken = googleAccount.authentication.idToken;
    if (idToken == null) {
      throw const AuthConfigurationException();
    }

    final credential = GoogleAuthProvider.credential(idToken: idToken);
    await _firebaseAuth.signInWithCredential(credential);
  }

  @override
  Future<void> signOut() async {
    await _firebaseAuth.signOut();
    await _googleSignInInitialized;
    await _googleSignIn.signOut();
  }
}

final class AuthConfigurationException implements Exception {
  const AuthConfigurationException();
}
